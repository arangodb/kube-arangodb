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

pipeline {
    options {
        buildDiscarder(logRotator(daysToKeepStr: '7', numToKeepStr: '10'))
    }
    agent any
    parameters {
      string(name: 'KUBECONFIG', defaultValue: '/home/jenkins/.kube/scw-183a3b', description: 'KUBECONFIG controls which k8s cluster is used', )
      string(name: 'TESTNAMESPACE', defaultValue: 'jenkins', description: 'TESTNAMESPACE sets the kubernetes namespace to ru tests in (this must be short!!)', )
    }
    stages {
        stage('Build') {
            steps {
                timestamps {
                    withEnv([
                    "IMAGETAG=${env.GIT_COMMIT}",
                    ]) {
                        sh "make"
                    }
                }
            }
        }
        stage('Test') {
            steps {
                timestamps {
                    lock("${params.TESTNAMESPACE}-${env.GIT_COMMIT}") {
                        withEnv([
                        "KUBECONFIG=${params.KUBECONFIG}",
                        "TESTNAMESPACE=${params.TESTNAMESPACE}-${env.GIT_COMMIT}",
                        "IMAGETAG=${env.GIT_COMMIT}",
                        "PUSHIMAGES=1",
                        ]) {
                            sh "make run-tests"
                        }
                    }
                }
            }
        }
    }

    post {
        failure {
            notifySlack('FAILURE')
        }

        success {
            notifySlack('SUCCESS')
        }
    }
}
