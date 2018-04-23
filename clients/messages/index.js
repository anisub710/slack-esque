var baseURL = "https://api.ask710.me/";

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
    new_user.password = $("sign_password").value;;
    new_user.passwordConf = $("password_conf").value;;
    var json = JSON.stringify(new_user)
    console.log(json)
}
