var baseURL = "https://api.ask710.me/";


//ADD UPLOAD PHOTO FUNCTIONALITY 
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
        getImage(image)        
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
        sendPhoto(photoAction.files, image);
    }
    actionDiv.appendChild(photoLink)

    card.appendChild(imageDiv);
    card.appendChild(content);
    card.appendChild(actionDiv)
    resultsRow.appendChild(card)
    results.appendChild(resultsRow);
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
}

function showLogin(){
    signDiv.classList.add("hidden");
    loginDiv.classList.remove("hidden");
    linkLogin.classList.add("active");
    linkSign.classList.remove("active");
    results.classList.add("hidden");     
}

function showError(error) {
    loginDiv.classList.add("hidden");
    signDiv.classList.add("hidden");
    results.classList.remove("hidden");   
    results.innerText = "Error: " + error
    results.classList.add("error")
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

function sendPhoto(files, image) {    
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
            console.log(response)
            getImage(image)
        }     
        return response.text().then((t) => Promise.reject(t))                               
    }).catch(function(error) {            
        showError(error)
    });    
}

function getImage(image){
    fetch(baseURL + "v1/users/me/avatar", {
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

