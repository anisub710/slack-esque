var baseURL = "https://api.ask710.me/";


//ADD UPLOAD PHOTO FUNCTIONALITY 
var $ = function(id) { return document.getElementById(id) };
var linkSign = document.getElementById("link_sign");
var linkLogin = document.getElementById("link_login");
var signDiv = document.getElementById("sign_up");
var loginDiv = document.getElementById("login");
var submitSign = document.getElementById("submit_sign");
var submitLogin = document.getElementById("submit_login")
linkSign.onclick = function(){
    signDiv.classList.remove("hidden");
    loginDiv.classList.add("hidden");
    linkSign.classList.add("active");
    linkLogin.classList.remove("active");
}

linkLogin.onclick = function(){
    signDiv.classList.add("hidden");
    loginDiv.classList.remove("hidden");
    linkLogin.classList.add("active");
    linkSign.classList.remove("active");
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
        console.log(response.headers.get('Authorization'))              
            return response.json()       
    }).then(function(data){        
        convertData(data)
    }).catch(function(error) {            
        console.log(error)
    });
}

function convertData(data) {

    console.log(data)
}

submitLogin.onclick = function() {
    var credentials = {};
    credentials.email = $("login_email").value;;
    credentials.password = $("login_password").value;;    
    // fetch(baseURL + "v1/sessions", {
    //     method: 'POST',
    //     body: new_user,
    //     headers: new Headers({
    //         'Content-Type': 'application/json',
    //         'Authorization': 'Bearer '            
    //     })
    // }).then(res => console.log(res))

}