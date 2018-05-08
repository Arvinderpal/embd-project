node {
    try{        
        ws("${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}/") {
            withEnv(["GOPATH=${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}"]) {
                env.PATH="${GOPATH}/bin:$PATH"
                environment {
                  // FIXME: this does not work ...
                  PROJECTDIR = "src/github.com/Arvinderpal/embd-project"
                }

                stage('Checkout'){
                    echo '###Checking out SCM###'
                    dir('src/github.com/Arvinderpal/embd-project') {
                      checkout scm
                    }
                }
                
                stage('Pre Test'){
                    echo '---Pulling Dependencies---'
                    sh 'go version'
                    sh 'go get -u github.com/golang/lint/golint'
                    //sh 'go get github.com/tebeka/go2xunit'
                }
        
                stage('Test'){    
                    // List all our project files
                    // Push our project files relative to ./src
                    sh 'cd src/github.com/Arvinderpal/embd-project && go list ./... | grep -v /vendor/ > projectPaths'
                    
                    // Use awk to concat `./src/` to start of every import path reported by go list and to replace newlines with spaces.
                    def paths = sh returnStdout: true, script: """awk '{printf "./src/%s ",\$0} END {print ""}' ./src/github.com/Arvinderpal/embd-project/projectPaths"""                
                  
                    echo '~~~Vetting~~~'
                    sh """go tool vet ${paths}"""

                    echo '~~~Linting~~~'
                    sh """golint ${paths}"""
                  
                    echo '~~~Testing~~~'
                    //sh """go test -race -cover ${paths}"""
                    sh """go test ${paths}"""
                    
                }
            
                stage('Build'){
                  dir('src/github.com/Arvinderpal/embd-project') {
                    echo '+++Building Executable+++'
                
                    //Produced binary is $GOPATH/<name>
                    sh """go build -ldflags '-s'"""
                  }
                }
            }
        }
    }catch (e) {
        // If there was an exception thrown, the build failed
        currentBuild.result = "FAILED"
        echo 'FAILED'
    } finally {
       
        def bs = currentBuild.result ?: 'SUCCESSFUL'
        if(bs == 'SUCCESSFUL'){
            echo 'SUCCESSFUL'
        }
    }
}

