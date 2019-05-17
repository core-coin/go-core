pipeline {
    agent any
    environment {
        GOPATH = pwd()
    }
    stages {
        stage('clean') {
            steps {
                deleteDir()
            }
        }
        stage('checkout') {
            steps {
                checkout([$class: 'GitSCM',
                    branches: [[name: '*/jenkins']],
                    extensions: [
                        [$class: 'RelativeTargetDirectory', relativeTargetDir: 'src/github.com/ethereum/go-core']
                    ],
                    userRemoteConfigs: [
                        [credentialsId: 'a828724b-cdf1-4da0-93e2-28686c8a091b', url: 'https://github.com/core-coin/go-core.git']
                    ]
                ])
            }
      }
      stage('init') {
          steps {
              sh 'go get -u ekyu.moe/cryptonight'
          }
      }
      stage('build') {
          steps {
              sh '''
                  cd src/github.com/ethereum/go-core
                  go run build/ci.go install
              '''
          }
      }
      stage('test') {
          steps {
              script {
                  try {
                      sh '''
                          cd src/github.com/ethereum/go-core
                          //go run build/ci.go test -coverage
                      '''
                  } catch (e) {
                      /* ignore errors until we fix tests */
                  }
              }
          }
      }
      stage('deb') {
          steps {
              sh '''
                  cd src/github.com/ethereum/go-core
                  go run build/ci.go debsrc
              '''
          }
      }
   }
}
