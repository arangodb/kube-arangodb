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
    def options = sh(returnStdout: true, script: "git log --reverse master..HEAD | grep -o '\[ci[^\[]*\]' | sed -E 's/\[ci (.*)\]/\1/'").trim().split("\n")
    for (opt in options) {
        def idx = opt.indexOf('=');
        if (idx > 0) {
            def key = opt.subString(0, idx);
            def value = opt.subString(idx+1);
            params[key] = value;
            println("Overwriting params.${key} with ${value}");
        }
    }
}

def kubeConfigRoot = "/home/jenkins/.kube"

def buildTestSteps(String kubeConfigRoot, String kubeconfig) {
    return {
        timestamps {
            withCredentials([string(credentialsId: 'ENTERPRISEIMAGE', variable: 'DEFAULTENTERPRISEIMAGE')]) { 
                withEnv([
                "DEPLOYMENTNAMESPACE=${params.TESTNAMESPACE}-${env.GIT_COMMIT}",
                "DOCKERNAMESPACE=${params.DOCKERNAMESPACE}",
                "ENTERPRISEIMAGE=${params.ENTERPRISEIMAGE}",
                "IMAGETAG=jenkins-test",
                "KUBECONFIG=${kubeConfigRoot}/${kubeconfig}",
                "LONG=${params.LONG ? 1 : 0}",
                "TESTOPTIONS=${params.TESTOPTIONS}",
                ]) {
                    sh "make run-tests"
                }
            }
        }
    }
}

def buildCleanupSteps(String kubeConfigRoot, String kubeconfig) {
    return {
        timestamps {
            withEnv([
                "DEPLOYMENTNAMESPACE=${params.TESTNAMESPACE}-${env.GIT_COMMIT}",
                "DOCKERNAMESPACE=${params.DOCKERNAMESPACE}",
                "KUBECONFIG=${kubeConfigRoot}/${kubeconfig}",
            ]) {
                sh "make cleanup-tests"
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
      string(name: 'ENTERPRISEIMAGE', defaultValue: '', description: 'ENTERPRISEIMAGE sets the docker image used for enterprise tests)', )
    }
    stages {
        stage("Prepare") {
            steps {
                script {
                    fetchParamsFromGitLog()
                }
            }
        }
        stage('Build') {
            steps {
                timestamps {
                    withEnv([
                    "DEPLOYMENTNAMESPACE=${params.TESTNAMESPACE}-${env.GIT_COMMIT}",
                    "DOCKERNAMESPACE=${params.DOCKERNAMESPACE}",
                    "IMAGETAG=jenkins-test",
                    "LONG=${params.LONG ? 1 : 0}",
                    "TESTOPTIONS=${params.TESTOPTIONS}",
                    ]) {
                        sh "make"
                        sh "make run-unit-tests"
                        sh "make docker-test"
                    }
                }
            }
        }
        stage('Test') {
            steps {
                script {
                    def configs = "${params.KUBECONFIGS}".split(",")
                    def testTasks = [:]
                    for (kubeconfig in configs) {
                        testTasks["${kubeconfig}"] = buildTestSteps(kubeConfigRoot, kubeconfig)
                    }
                    parallel testTasks
                }
            }
        }
    }

    post {
        always {
            script {
                def configs = "${params.KUBECONFIGS}".split(",")
                def cleanupTasks = [:]
                for (kubeconfig in configs) {
                    cleanupTasks["${kubeconfig}"] = buildCleanupSteps(kubeConfigRoot, kubeconfig)
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
