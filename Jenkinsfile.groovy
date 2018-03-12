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
      booleanParam(name: 'LONG', defaultValue: false, description: 'Execute long running tests')
      string(name: 'KUBECONFIG', defaultValue: '/home/jenkins/.kube/scw-183a3b', description: 'KUBECONFIG controls which k8s cluster is used', )
      string(name: 'DOCKERNAMESPACE', defaultValue: 'arangodb', description: 'DOCKERNAMESPACE sets the docker registry namespace in which the operator docker image will be pushed', )
      string(name: 'TESTNAMESPACE', defaultValue: 'jenkins', description: 'TESTNAMESPACE sets the kubernetes namespace to ru tests in (this must be short!!)', )
      string(name: 'ENTERPRISEIMAGE', defaultValue: '', description: 'ENTERPRISEIMAGE sets the docker image used for enterprise tests)', )
    }
    stages {
        stage('Build') {
            steps {
                timestamps {
                    withEnv([
                    "IMAGETAG=${env.GIT_COMMIT}",
                    "DOCKERNAMESPACE=${params.DOCKERNAMESPACE}",
                    ]) {
                        sh "make"
                        sh "make run-unit-tests"
                    }
                }
            }
        }
        stage('Test') {
            steps {
                timestamps {
                    lock("${params.TESTNAMESPACE}-${env.GIT_COMMIT}") {
                        withCredentials([string(credentialsId: 'ENTERPRISEIMAGE', variable: 'DEFAULTENTERPRISEIMAGE')]) { 
                            withEnv([
                            "DOCKERNAMESPACE=${params.DOCKERNAMESPACE}",
                            "ENTERPRISEIMAGE=${params.ENTERPRISEIMAGE}",
                            "IMAGETAG=${env.GIT_COMMIT}",
                            "KUBECONFIG=${params.KUBECONFIG}",
                            "LONG=${params.LONG ? 1 : 0}",
                            "DEPLOYMENTNAMESPACE=${params.TESTNAMESPACE}-${env.GIT_COMMIT}",
                            ]) {
                                sh "make run-tests"
                            }
                        }
                    }
                }
            }
        }
    }

    post {
        always {
            timestamps {
                withEnv([
                    "DOCKERNAMESPACE=${params.DOCKERNAMESPACE}",
                    "KUBECONFIG=${params.KUBECONFIG}",
                    "DEPLOYMENTNAMESPACE=${params.TESTNAMESPACE}-${env.GIT_COMMIT}",
                ]) {
                    sh "make cleanup-tests"
                }
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
