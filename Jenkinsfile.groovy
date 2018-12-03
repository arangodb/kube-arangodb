def notifySlack(String buildStatus = 'STARTED') {
    // Build status of null means success.
    buildStatus = buildStatus ?: 'SUCCESS'

    def color

    if (buildStatus == 'STARTED') {
        color = '#D4DADF'
    } else if (buildStatus == 'SUCCESS') {
        color = '#BDFFC3'
    } else if (buildStatus == 'UNSTABLE') {
        color = '#FFFE89'
    } else {
        color = '#FF9FA1'
    }

    def msg = "${buildStatus}: `${env.JOB_NAME}` #${env.BUILD_NUMBER}: ${env.GIT_COMMIT}\n${env.BUILD_URL}"

    slackSend(color: color, channel: '#status-k8s', message: msg)
}

def fetchParamsFromGitLog() {
    def myParams = [:];
    // Copy configured params 
    for (entry in params) {
        myParams[entry.key] = entry.value;
    }

    // Is this a LONG test?
    if ("${env.JOB_NAME}" == "kube-arangodb-long") {
        myParams["LONG"] = true;
        myParams["KUBECONFIGS"] = "gke-jenkins-1";
    }

    // Fetch params configured in git commit messages 
    // Syntax: [ci OPT=value]
    // Example: [ci TESTOPTIONS="-test.run ^TestSimpleSingle$"]
    def options = sh(returnStdout: true, script: "git log --reverse remotes/origin/master..HEAD | grep -o \'\\[ci[^\\[]*\\]\' | sed -E \'s/\\[ci (.*)\\]/\\1/\'").trim().split("\n")
    for (opt in options) {
        def idx = opt.indexOf('=');
        if (idx > 0) {
            def key = opt.substring(0, idx);
            def value = opt.substring(idx+1).replaceAll('^\"|\"$', '');
            myParams[key] = value;
            //println("Overwriting myParams.${key} with ${value}");
        }
    }

    // Show params in log
    for (entry in myParams) {
        println("Using myParams.${entry.key} with ${entry.value}");
    }

    return myParams;
}

def kubeConfigRoot = "/home/jenkins/.kube"

def buildBuildSteps(Map myParams) {
    return {
        timestamps {
            timeout(time: 15) {
                withEnv([
                "DEPLOYMENTNAMESPACE=${myParams.TESTNAMESPACE}-${env.GIT_COMMIT}",
                "DOCKERNAMESPACE=${myParams.DOCKERNAMESPACE}",
                "IMAGETAG=jenkins-test",
                "LONG=${myParams.LONG ? 1 : 0}",
                "TESTOPTIONS=${myParams.TESTOPTIONS}",
                ]) {
                    sh "make clean"
                    sh "make"
                    sh "make run-unit-tests"
                    sh "make docker-test"
                }
            }
        }
    }
}

def buildTestSteps(Map myParams, String kubeConfigRoot, String kubeconfig) {
    return {
        timestamps {
            timeout(time: myParams.LONG ? 180 : 30) {
                withCredentials([
                    string(credentialsId: 'ENTERPRISEIMAGE', variable: 'DEFAULTENTERPRISEIMAGE'),
                    string(credentialsId: 'ENTERPRISELICENSE', variable: 'DEFAULTENTERPRISELICENSE'),
                ]) { 
                    withEnv([
                    "CLEANDEPLOYMENTS=1",
                    "DEPLOYMENTNAMESPACE=${myParams.TESTNAMESPACE}-${env.GIT_COMMIT}",
                    "DOCKERNAMESPACE=${myParams.DOCKERNAMESPACE}",
                    "ENTERPRISEIMAGE=${myParams.ENTERPRISEIMAGE}",
                    "ENTERPRISELICENSE=${myParams.ENTERPRISELICENSE}",
                    "ARANGODIMAGE=${myParams.ARANGODIMAGE}",
                    "IMAGETAG=jenkins-test",
                    "KUBECONFIG=${kubeConfigRoot}/${kubeconfig}",
                    "LONG=${myParams.LONG ? 1 : 0}",
                    "TESTOPTIONS=${myParams.TESTOPTIONS}",
                    ]) {
                        sh "make run-tests"
                    }
                }
            }
        }
    }
}

def buildCleanupSteps(Map myParams, String kubeConfigRoot, String kubeconfig) {
    return {
        timestamps {
            timeout(time: 15) {
                withEnv([
                    "DEPLOYMENTNAMESPACE=${myParams.TESTNAMESPACE}-${env.GIT_COMMIT}",
                    "DOCKERNAMESPACE=${myParams.DOCKERNAMESPACE}",
                    "KUBECONFIG=${kubeConfigRoot}/${kubeconfig}",
                ]) {
                    sh "./scripts/collect_logs.sh ${env.DEPLOYMENTNAMESPACE} ${kubeconfig}" 
                    archive includes: 'logs/*'
                    sh "make cleanup-tests"
                }
            }
        }
    }
}

pipeline {
    options {
        buildDiscarder(logRotator(daysToKeepStr: '7', numToKeepStr: '10'))
        lock resource: 'kube-arangodb'
    }
    agent any
    parameters {
      booleanParam(name: 'LONG', defaultValue: false, description: 'Execute long running tests')
      string(name: 'DOCKERNAMESPACE', defaultValue: 'arangodb', description: 'DOCKERNAMESPACE sets the docker registry namespace in which the operator docker image will be pushed', )
      string(name: 'KUBECONFIGS', defaultValue: 'kube-ams1,scw-183a3b', description: 'KUBECONFIGS is a comma separated list of Kubernetes configuration files (relative to /home/jenkins/.kube) on which the tests are run', )
      string(name: 'TESTNAMESPACE', defaultValue: 'jenkins', description: 'TESTNAMESPACE sets the kubernetes namespace to ru tests in (this must be short!!)', )
      string(name: 'ENTERPRISEIMAGE', defaultValue: '', description: 'ENTERPRISEIMAGE sets the docker image used for enterprise tests', )
      string(name: 'ARANGODIMAGE', defaultValue: '', description: 'ARANGODIMAGE sets the docker image used for tests (except enterprise and update tests)', )
      string(name: 'ENTERPRISELICENSE', defaultValue: '', description: 'ENTERPRISELICENSE sets the enterprise license key for enterprise tests', )
      string(name: 'TESTOPTIONS', defaultValue: '', description: 'TESTOPTIONS is used to pass additional test options to the integration test', )
    }
    stages {
        stage('Build') {
            steps {
                script {
                    def myParams = fetchParamsFromGitLog();
                    def buildSteps = buildBuildSteps(myParams);
                    buildSteps();
                }
            }
        }
        stage('Test') {
            steps {
                script {
                    def myParams = fetchParamsFromGitLog();
                    def configs = "${myParams.KUBECONFIGS}".split(",")
                    def testTasks = [:]
                    for (kubeconfig in configs) {
                        testTasks["${kubeconfig}"] = buildTestSteps(myParams, kubeConfigRoot, kubeconfig)
                    }
                    parallel testTasks
                }
            }
        }
    }

    post {
        always {
            script {
                def myParams = fetchParamsFromGitLog();
                def configs = "${myParams['KUBECONFIGS']}".split(",")
                def cleanupTasks = [:]
                for (kubeconfig in configs) {
                    cleanupTasks["${kubeconfig}"] = buildCleanupSteps(myParams, kubeConfigRoot, kubeconfig)
                }
                parallel cleanupTasks
            }
        }
        failure {
            notifySlack('FAILURE')
        }

        success {
            notifySlack('SUCCESS')
        }
    }
}
