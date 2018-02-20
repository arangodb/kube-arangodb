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
      string(name: 'TESTNAMESPACE', defaultValue: 'arangodb-operator-tests', description: 'TESTNAMESPACE sets the kubernetes namespace to ru tests in', )
    }
    stages {
        stage('Build') {
            steps {
                timestamps {
                    sh "make"
                }
            }
        }
        stage('Test') {
            steps {
                timestamps {
                    lock("kubernetes-operator-tests") {
                        withEnv([
                        'TESTNAMESPACE=${params.TESTNAMESPACE}',
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
