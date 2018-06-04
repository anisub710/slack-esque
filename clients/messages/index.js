var baseURL = "https://api.ask710.me/";


var $ = function(id) { return document.getElementById(id) };
var linkSign = document.getElementById("link_sign");
var linkLogin = document.getElementById("link_login");
var signDiv = document.getElementById("sign_up");
var loginDiv = document.getElementById("login");
var submitSign = document.getElementById("submit_sign");
var submitLogin = document.getElementById("submit_login");
var results = document.getElementById("results");
var resultsRow = document.getElementById("results_row");
var myStorage = window.localStorage
var search = document.getElementById("search");
var searchForm = document.getElementById("search-form");
var timer;
var suggestions = document.getElementById("list");
var typingPause = 700;

search.addEventListener("keyup", function(){
    clearTimeout(timer);
    timer = setTimeout(function(){
      queryUsers(search)
    },typingPause);
  
  
});
  search.addEventListener("keydown", function(e){
    clearTimeout(timer);    
    suggestions.innerHTML = "";
  
});

function queryUsers(input){
    if(input.value != "") {
        fetch(baseURL + "v1/users?q=" + input.value, {           
            headers: new Headers({            
                'Authorization': myStorage.getItem("sessionID")            
            })
        }).then(function(response){  
            if(response.status < 300){                         
                return response.json()      
            }     
            return response.text().then((t) => Promise.reject(t))                               
        }).then(function(data){                
            // console.log(data)
            searchUsers(data)
        }).catch(function(error) {            
            showError(error)
        });                
    }
    
}

function searchUsers(response){    
    if(response != null){
        for(var i = 0; i < response.length; i++){        
            data = response[i]
            suggestionWord = document.createElement("div");
            specificWord = document.createElement("span")
            specificWord.classList.add("specific")
            specificWord.innerHTML = data.firstName + " " + data.lastName 
            + " (@" + data.userName + ")";
            suggestionWord.classList.add("word");           
            var image = document.createElement("img");  
            image.classList.add("profile")  
            if(data.photoURL.includes("gravatar")){
                image.src = data.photoURL  
            }else{
                getImage(image, data.id)        
            }
            suggestionWord.appendChild(image);
            suggestionWord.appendChild(specificWord);
            suggestions.appendChild(suggestionWord);
        }
    }
}



linkSign.onclick = function(){
    showSignUp();
}

linkLogin.onclick = function(){
    showLogin();
}

submitSign.onclick = function() {
    var new_user = {};
    new_user.firstName = $("first_name").value;
    new_user.lastName = $("last_name").value;
    new_user.userName = $("user_name").value;
    new_user.email = $("sign_email").value;;
    new_user.password = $("sign_password").value;
    new_user.passwordConf = $("password_conf").value;   
    console.log(new_user) 
    fetch(baseURL + "v1/users", {
        method: 'post',
        body: JSON.stringify(new_user),        
        headers: new Headers({
            'Content-Type': 'application/json'            
        })
    }).then(function(response){  
        if(response.status < 300){
            myStorage.setItem('sessionID', response.headers.get("Authorization"))
            console.log(response.headers.get("Authorization"))
            return response.json()      
        }     
        return response.text().then((t) => Promise.reject(t))                               
    }).then(function(data){        
        convertData(data)
        console.log(data)
    }).catch(function(error) {            
        showError(error);
    });
}

function convertData(data) {
    results.classList.remove("error");
    results.innerText = "";
    var welcome = document.createElement("h3");
    welcome.innerText = "Welcome"
    welcome.classList.add("header");
    results.appendChild(welcome);
    resultsRow.innerHTML = "";
    loginDiv.classList.add("hidden");
    signDiv.classList.add("hidden");
    results.classList.remove("hidden");
    searchForm.classList.remove("hidden");

    var card = document.createElement("div");
    card.classList.add("card");

    var content = document.createElement("div");
    content.classList.add("card-content");    

    var title = document.createElement("span");
    title.classList.add("card-title")
    title.innerText = data.firstName + " " + data.lastName;

    content.appendChild(title);      

    var imageDiv = document.createElement("div");
    imageDiv.classList.add("card-image");  
    var image = document.createElement("img");  
    image.id = "test"  
    if(data.photoURL.includes("gravatar")){
        image.src = data.photoURL  
    }else{
        getImage(image, data.id)        
    }
    
    imageDiv.appendChild(image)     

    
    var actionDiv = document.createElement("div");
    actionDiv.classList.add("card-action")
    var action = document.createElement("a")
    action.innerText = "Sign Out"
    action.onclick = function() {
        signOut()
    }
    actionDiv.appendChild(action)   

    var photoLink = document.createElement("a")
    photoLink.innerText = "Upload Photo"
    var photoAction = document.createElement("input")
    photoAction.type = "file"
    photoAction.accept = "image/*"
    photoAction.id = "upload-pic"
    photoLink.onclick = function() {        
        photoAction.click();        
    }    
    photoAction.onchange = function() {
        sendPhoto(photoAction.files, image, data.id);
    }
    actionDiv.appendChild(photoLink)

    card.appendChild(imageDiv);
    card.appendChild(content);
    card.appendChild(actionDiv)
    resultsRow.appendChild(card)
    results.appendChild(resultsRow);


    const host = "api.ask710.me";

    // const status = document.querySelector("#status")
    // const notifications = document.querySelector("#notifications");
    // const errors = document.querySelector("#errors");

    //use `wss://` if you are connecting to an HTTPS server

    const websocket = new WebSocket("wss://" + host + "/v1/ws?auth=" + myStorage.getItem("sessionID"));
    websocket.addEventListener("error", function(err) {
        console.log("Error: "+ err.message);
    });
    websocket.addEventListener("open", function() {
        console.log("Status: Open");
    });
    websocket.addEventListener("close", function() {
        console.log("Status: Closed");
    });
    websocket.addEventListener("message", function(event) {
        console.log("Notification: " + event.data);
        let p = document.createElement("p");        
    });

    document.querySelector("#createchannel").addEventListener("click", function() {
        let channel = {name: "please 23", private: false};
        fetch("https://" + host + "/v1/channels", {
            method: 'POST',
            body: JSON.stringify(channel),
            headers: new Headers({
                'Content-Type': 'application/json',
                'Authorization': myStorage.getItem("sessionID")            
            })
        }).then(function(response){
            if (response.status < 300){
                return response.json();
            }                        
            return response.text().then((t) => Promise.reject(t)) 
        }).then(function(data){
            console.log(data);
        })
        .catch(function(err) {
            console.log("Error: " + err);
        });
    });
    document.querySelector("#updatechannel").addEventListener("click", function() {
        let channel = {name: "changed12345678", description: "changed description"};
        fetch("https://" + host + "/v1/channels/10", {
            method: 'PATCH',
            body: JSON.stringify(channel),
            headers: new Headers({
                'Content-Type': 'application/json',
                'Authorization': myStorage.getItem("sessionID")            
            })
        }).then(function(response){
            if (response.status < 300){
                return response.json();
            }                        
            return response.text().then((t) => Promise.reject(t))                               
        }).then(function(data){
            console.log(data);
        })
        .catch(function(err) {
            console.log("Error: " + err);
        });
    });
    document.querySelector("#deletechannel").addEventListener("click", function() {        
        fetch("https://" + host + "/v1/channels/5", {
            method: 'DELETE',            
            headers: new Headers({
                'Content-Type': 'application/json',
                'Authorization': myStorage.getItem("sessionID")            
            })
        }).then(function(response){
            if (response.status < 300){
                return response
            }                        
            return response.text().then((t) => Promise.reject(t))                               
        }).then(function(data){
            console.log(data);
        })
        .catch(function(err) {
            console.log("Error: " + err);
        });
    });
    document.querySelector("#createmessage").addEventListener("click", function() {    
        let message = {body: "hello"};    
        fetch("https://" + host + "/v1/channels/10", {
            method: 'POST',     
            body: JSON.stringify(message),       
            headers: new Headers({
                'Content-Type': 'application/json',
                'Authorization': myStorage.getItem("sessionID")            
            })
        }).then(function(response){
            if (response.status < 300){
                return response.json();
            }                        
            return response.text().then((t) => Promise.reject(t))                               
        }).then(function(data){
            console.log(data);
        })
        .catch(function(err) {
            console.log("Error: " + err);
        });
    });

    document.querySelector("#updatemessage").addEventListener("click", function() {    
        let message = {body: "not hello"};    
        fetch("https://" + host + "/v1/messages/1", {
            method: 'PATCH',     
            body: JSON.stringify(message),       
            headers: new Headers({
                'Content-Type': 'application/json',
                'Authorization': myStorage.getItem("sessionID")            
            })
        }).then(function(response){
            if (response.status < 300){
                return response.json();
            }                        
            return response.text().then((t) => Promise.reject(t))                               
        }).then(function(data){
            console.log(data);
        })
        .catch(function(err) {
            console.log("Error: " + err);
        });
    });
    document.querySelector("#deletemessage").addEventListener("click", function() {             
        fetch("https://" + host + "/v1/messages/1", {
            method: 'DELETE',     
            headers: new Headers({
                'Content-Type': 'application/json',
                'Authorization': myStorage.getItem("sessionID")            
            })
        }).then(function(response){
            if (response.status < 300){
                return response;
            }                        
            return response.text().then((t) => Promise.reject(t))                               
        }).then(function(data){
            console.log(data);
        })
        .catch(function(err) {
            console.log("Error: " + err);
        });
    });

}

submitLogin.onclick = function() {
    var credentials = {};
    credentials.email = $("login_email").value;;
    credentials.password = $("login_password").value;;    
    fetch(baseURL + "v1/sessions", {
        method: 'POST',
        body: JSON.stringify(credentials),
        headers: new Headers({
            'Content-Type': 'application/json',
            'Authorization': myStorage.getItem("sessionID")            
        })
    }).then(function(response){  
        if(response.status < 300){  
            myStorage.setItem('sessionID', response.headers.get("Authorization"))          
            return response.json()      
        }     
        return response.text().then((t) => Promise.reject(t))                               
    }).then(function(data){        
        convertData(data)
        console.log(data)
    }).catch(function(error) {            
        showError(error)
    });

}

function showSignUp(){
    signDiv.classList.remove("hidden");
    loginDiv.classList.add("hidden");
    linkSign.classList.add("active");
    linkLogin.classList.remove("active");
    results.classList.add("hidden");
    searchForm.classList.add("hidden");
}

function showLogin(){
    signDiv.classList.add("hidden");
    loginDiv.classList.remove("hidden");
    linkLogin.classList.add("active");
    linkSign.classList.remove("active");
    results.classList.add("hidden");   
    searchForm.classList.add("hidden");     
}

function showError(error) {
    loginDiv.classList.add("hidden");
    signDiv.classList.add("hidden");
    results.classList.remove("hidden");   
    results.innerText = "Error: " + error
    results.classList.add("error");
    searchForm.classList.add("hidden");   
}

function signOut(){
    fetch(baseURL + "v1/sessions/mine", {
        method: 'DELETE',
        headers: new Headers({            
            'Authorization': myStorage.getItem("sessionID") 
        })
        }) .then(function(response){
            if(response.status < 300){
                console.log(response)
                myStorage.removeItem("sessionID")
                showLogin(); 
                M.toast({html: 'Signed Out'})
                return null
            }
            return response.text().then((t) => Promise.reject(t))                               
        }).catch(function(error) {                     
            showError(error)
        });    
}

function sendPhoto(files, image, id) {    
    var formData = new FormData()
    console.log(files[0])
    formData.append("avatar", files[0])
    fetch(baseURL + "v1/users/me/avatar", {
        method: 'PUT',
        body: formData,
        headers: new Headers({            
            'Authorization': myStorage.getItem("sessionID")            
        })       
    }).then(function(response){  
        if(response.status < 300){
            getImage(image, id)
            return null
        }     
        return response.text().then((t) => Promise.reject(t))                               
    }).catch(function(error) {            
        showError(error)
    });    
}

function getImage(image, id){
    fetch(baseURL + "v1/users/" + id + "/avatar", {
        headers: new Headers({            
            'Authorization': myStorage.getItem('sessionID')            
        })   
    }).then(function(response){
        if(response.ok){
            return response.arrayBuffer()
        }  
        return response.text().then((t) => Promise.reject(t))                                     
    }).then(function(data){
        var base64Flag = 'data:image/jpeg;base64,';
        var imageStr = arrayBufferToBase64(data);
        image.src = base64Flag + imageStr;
    }).catch(function(error) {            
        showError(error)
    });
}

function arrayBufferToBase64(buffer) {
    var binary = '';
    var bytes = [].slice.call(new Uint8Array(buffer));
  
    bytes.forEach((b) => binary += String.fromCharCode(b));
  
    return window.btoa(binary);
  };




