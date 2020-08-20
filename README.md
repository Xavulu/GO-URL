![Go URL logo](/images/go_url.png)

# GO URL 
>GO URL is a url shortening api made with go gin/gonic and postgres. I made this just as a simple excursion into backend development and api design. As such there are no plans to actually host this/build a full web app around it at the moment however anyone is free to use/modify this project for their own use. 

## Installation 
As a prerequisite to running this application you need to have go, docker and docker-compose installed on your pc as a minimum and postgres if you want to have a local database rather than one in a docker container. 

```
MacOS setup, run the following commands in your terminal:  

    brew update 
    brew install docker 
    brew install docker-machine 
    brew cask install virtualbox 
    brew install docker-compose
    brew install golang <-- theres more you have to setup after this but I wont go over that here
``` 

For Linux docker/docker-compose setup, read the official docs [here](https://docs.docker.com/engine/install/ubuntu/) and [here](https://docs.docker.com/compose/install/). Then run the following commands to install golang if you haven't already: 
``` 
    curl -O https://storage.googleapis.com/golang/go1.12.9.linux-amd64.tar.gz 
    sha256sum go1.12.9.linux-amd64.tar.gz <-- use this to verify 
    tar -xvf go1.12.9.linux-amd64.tar.gz 
    sudo chown -R root:root ./go
    sudo mv go /usr/local 

    then add go to your path
``` 
After having everything set up clone the repository to your pc 

## Running the app
Navigate to the project directory and run the following commands: 
```
    docker-compose build  
    docker-compose up 

    to run in background use docker-compose up -d instead and
    shutdown with docker-compose stop
``` 
This api runs on localhost:8080 and its postgres database is bound to port 8001 on your local pc. To connect to the databse from within its container run this command: 
``` psql postgres://short:777777@lulu:5432/shorturl ``` 
to connect to the database from outside the container use this instead: 

 ``` psql postgres://short:777777@localhost:8001/shorturl ``` 

The api has four endpoints: 
- POST http://localhost:8080/entry <- short url creation
- GET http://localhost:8080/o/:id <- redirection to original url
- GET http://localhost:8080/view/:id <- view original url without redirection
- GET http://localhost:8080/stats/:id <- visit times for your short url 

To generate shorturl send a post request with curl like this: 
```  
    curl -d '{"url" : "https://www.google.com"}' -H "Content-Type: application/json" -X POST http://localhost:8080/entry 
``` 
Which returns the following response: 
``` 
    {"code":200,
    "message":
        {"id":2,
        "original_url" : "https://www.google.com",
        "short_url" : "http://localhost:8080/o/c",
        "entry_date" : "2020-08-19T20:01:28.1963621Z"}
    }
``` 
You can then enter your short url into your browser which redirects you to the original link. To view what a shorturl links to without actually clicking it make the following GET request with curl:
``` 
    curl http://localhost:8080/view/c
``` 
Which returns this response: 
``` 
    {"code":200,
    "message":"http://localhost:8080/o/c links to https://www.google.com"
    }
``` 
To view the times when your shorturl was used you can sent a get request like this: 
``` 
    get http://localhost:8080/stats/c
``` 
Which returns the following after your shorturl has been used 3 times: 
``` 
    {"code":200,
    "message":{
        "original_url":"https://www.google.com","short_url":"http://localhost:8080/o/c","visited":true,
        "visit_count":3,
        "visit_times":
            ["2020-08-19T22:58:50.590696Z",
            "2020-08-19T22:58:52.327455Z",
            "2020-08-19T22:58:53.785104Z"]
            }
    }
```
## TODOS

- Proper unit testing 
- Add logging to mongodb or some other service 
- Add swagger/openapi documentation 
- Frontend client (still wont be hosted because hosting and manging a url shortener comes with a lot of growing pains I am not interested in dealing with) 
- Integrate the google safe browsing api to prevent malicious from being entered in the database (a common problem for many url shortening services) 






