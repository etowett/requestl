#!groovy

// @Library('etowix_shared')
import groovy.json.JsonOutput

helm_chart_version = "0.1.0"
cron_helm_chart_version = "0.1.0"

pipeline {
    agent {
        kubernetes {
            label "build-requestl-${BUILD_NUMBER}"
            defaultContainer 'jnlp'
            yaml """
apiVersion: v1
kind: Pod
metadata:
  labels:
    some-label: "build-requestl-${BUILD_NUMBER}"
spec:
  containers:
  - name: runner
    image: 240603544178.dkr.ecr.eu-west-3.amazonaws.com/jenkins-slave:v0.0.1
    command:
    - cat
    tty: true
    volumeMounts:
    - mountPath: /var/run/docker.sock
      name: docker-socket
  volumes:
  - name: docker-socket
    hostPath:
      path: /var/run/docker.sock
      type: File
"""
    }
  }
    environment {
        GITHUB_ACCESS_TOKEN  = credentials('github-token')
    }

    stages {

        stage('Checkout Code') {
            steps {
                checkout scm
            }
        }

        stage('Build the deploy image') {
            steps {
                container('runner') {
                    sh """
                        docker build -t requestl .
                    """
                }
            }
        }

        stage('Login to ECR'){
            steps{
                container('runner') {
                    sh '$(aws ecr get-login --no-include-email --region eu-west-3)'
                }
            }
        }

        stage('Publish to ecr registry') {
            steps {
                container('runner') {
                    sh """
                        docker tag requestl 240603544178.dkr.ecr.eu-west-3.amazonaws.com/requestl:${env.BRANCH_NAME}-${GIT_COMMIT.take(10)}
                        docker push 240603544178.dkr.ecr.eu-west-3.amazonaws.com/requestl:${env.BRANCH_NAME}-${GIT_COMMIT.take(10)}
                    """
                }
            }
        }

        stage('Deploy service') {
            steps {
                container('runner') {
                    sh """
                        helm repo update
                        helm upgrade -i --debug requestl etowix/app \
                            --version ${helm_chart_version} \
                            --set image.tag=${env.BRANCH_NAME}-${GIT_COMMIT.take(10)} \
                            -f ./helm/${env.ENV}.yaml --namespace=${env.ENV}
                        kubectl rollout status deployment.apps/requestl --namespace=${env.ENV}
                    """
                }
            }
        }
    }
}
