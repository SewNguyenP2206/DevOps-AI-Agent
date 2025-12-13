pipeline {
    agent any

    environment {
        GOPATH = "${WORKSPACE}/go"
        PATH = "${env.GOPATH}/bin:${env.PATH}"
    }

    stages {
        stage('Checkout') {
            steps {
                git url: 'https://github.com/SewNguyenP2206/DevOps-AI-Agent.git', branch: 'main'
            }
        }

        stage('Setup Go') {
            steps {
                sh 'go version'
                sh 'go mod tidy'
            }
        }

        stage('Build') {
            steps {
                sh 'go build -o myapp ./...'
            }
        }

        stage('Test') {
            steps {
                sh 'go test ./...'
            }
        }

        stage('Archive') {
            steps {
                archiveArtifacts artifacts: 'myapp', fingerprint: true
            }
        }
    }

    post {
        success {
            echo 'Build thành công'
        }
        failure {
            echo 'Build thất bại'
        }
    }
}
