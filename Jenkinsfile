@Library('jenkins-shared-lib@4.18.1') _

pipeline {
    agent {
        kubernetes {
            yaml '''
                apiVersion: v1
                kind: Pod
                metadata:
                labels:
                    jenkins/jenkins-jenkins-agent: true
                    jenkins/label: terraform-provider-qbee
                    app.kubernetes.io/name: terraform-provider-qbee
                spec:
                    containers:
                    - name: golang
                      image: golang:1.21-bookworm
                      command:
                      - cat
                      tty: true
                    volumes:
                    - name: docker-socket
                      hostPath:
                        path: /var/run/docker.sock
            '''
            nodeSelector 'instance-type=spot'
        }
    }

    stages {
        stage('Compile') {
            steps {
                sh 'go build'
            }
        }

        stage('Test') {
            steps {
                sh 'go test ./...'
            }
        }

        stage ('Release') {
            when {
                buildingTag()
            }

            steps {
                sh 'curl -sfL https://goreleaser.com/static/run | bash'
            }
        }
    }
}
