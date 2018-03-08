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

def kubeConfigRoot = "/home/jenkins/.kube/"

def buildTestSteps(String kubeconfig) {
    return {
        timestamps {
            lock("${kubeconfig}-${params.TESTNAMESPACE}-${env.GIT_COMMIT}") {
                withCredentials([string(credentialsId: 'ENTERPRISEIMAGE', variable: 'DEFAULTENTERPRISEIMAGE')]) { 
                    withEnv([
                    "ENTERPRISEIMAGE=${params.ENTERPRISEIMAGE}",
                    "IMAGETAG=${env.GIT_COMMIT}",
                    "KUBECONFIG=${kubeConfigRoot}/${kubeconfig}",
                    "LONG=${params.LONG ? 1 : 0}",
                    "PUSHIMAGES=1",
                    "TESTNAMESPACE=${params.TESTNAMESPACE}-${env.GIT_COMMIT}",
                    ]) {
                        sh "make run-tests"
                    }
                }
            }
        }
    }
}

def buildCleanupSteps(String kubeconfig) {
    return {
        timestamps {
                withEnv([
                    "KUBECONFIG=${kubeConfigRoot}/${kubeconfig}",
                    "TESTNAMESPACE=${params.TESTNAMESPACE}-${env.GIT_COMMIT}",
                ]) {
                    sh "make cleanup-tests"
                }
            }
        }
    }
}

pipeline {
    options {
        buildDiscarder(logRotator(daysToKeepStr: '7', numToKeepStr: '10'))
    }
    agent any
    parameters {
      booleanParam(name: 'LONG', defaultValue: false, description: 'Execute long running tests')
      string(name: 'KUBECONFIGS', defaultValue: 'scw-183a3b,c11', description: 'KUBECONFIGS is a comma separated list of Kubernetes configuration files (relative to /home/jenkins/.kube) on which the tests are run', )
      string(name: 'TESTNAMESPACE', defaultValue: 'jenkins', description: 'TESTNAMESPACE sets the kubernetes namespace to ru tests in (this must be short!!)', )
      string(name: 'ENTERPRISEIMAGE', defaultValue: '', description: 'ENTERPRISEIMAGE sets the docker image used for enterprise tests)', )
    }
    stages {
        stage('Build') {
            steps {
                timestamps {
                    withEnv([
                    "IMAGETAG=${env.GIT_COMMIT}",
                    ]) {
                        sh "make"
                        sh "make run-unit-tests"
                    }
                }
            }
        }
        stage('Test') {
            steps {
                def configs = "{params.KUBECONFIGS}".split(",")
                def testTasks[:]
                for (kubeconfig in configs) {
                    testTasks["${kubeconfig}"] = buildTestSteps(kubeconfig)
                }
                parallel testTasks
            }
        }
    }

    post {
        always {
            def configs = "{params.KUBECONFIGS}".split(",")
            def cleanupTasks[:]
            for (kubeconfig in configs) {
                cleanupTasks["${kubeconfig}"] = buildCleanupSteps(kubeconfig)
            }
            parallel cleanupTasks
        }
        failure {
            notifySlack('FAILURE')
        }

        success {
            notifySlack('SUCCESS')
        }
    }
}
