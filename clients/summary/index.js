var baseURL = "https://localhost:4000/";
var resource = "v1/summary";
var param = "?url="
// var $ = function(id) { return document.getElementById(id) };

var form = document.getElementById("search-form");
var input = document.getElementById("search");
var results = document.getElementById("results");
var searchIcon = document.getElementById("search-icon");
form.addEventListener("submit", evt => {
    results.innerHTML = "";
    evt.preventDefault();   
    getMeta(evt);
});

searchIcon.onclick = function() {
    results.innerHTML = "";
    getMeta();
};

function getMeta(){        
    fetch(baseURL + resource + param + input.value)
        .then(function(response){            
            if(response.ok){
                return response.json();  
            }            
            return response.text().then((t) => Promise.reject(t))            
        }).then(function(data){
            fillCard(data);                     
        }).catch(function(error) {            
            showError(error);            
        });

}

//Add link from card to URL
function fillCard(data) {
    if(data != null){
        var card = document.createElement("div");
        card.classList.add("card");
        var content = document.createElement("div");
        content.classList.add("card-content");    
        var title = document.createElement("span");
        title.classList.add("card-title")
        title.innerText = data.title;
        var description = document.createElement("p");
        description.innerText = data.description;   
        if(data.videos != null) {
            
            if(data.videos[0].type != null && data.videos[0].type.startsWith("video/")){
                var video = document.createElement("video")
                video.src = data.videos[0].url;
                video.classList.add("my-image");
                card.appendChild(video);
            }else {
                var video = document.createElement("iframe")
                video.src = data.videos[0].url;
                video.classList.add("my-image");
                card.appendChild(video);
            }
            
        } else if(data.images != null){
            var carousel = document.createElement("div");
            carousel.classList.add("carousel");
            carousel.classList.add("carousel-slider");
            carousel.classList.add("center");
            for(var i = 0; i < data.images.length; i++){            
                var carouselItem = document.createElement("div");
                carouselItem.style.transition = "ease 0.5s"
                carouselItem.href = '"' + i + '!"';
                carouselItem.classList.add("carousel-item");
                // var imageDiv = document.createElement("div");
                // imageDiv.classList.add("card-image");
                var image = document.createElement("img");                
                image.src = data.images[i].url;
                // imageDiv.appendChild(image); 
                carouselItem.appendChild(image); 
                // carouselItem.appendChild(imageDiv);
                carousel.appendChild(carouselItem);
            }            
            card.appendChild(carousel);  
            // $('.carousel.carousel-slider').carousel({
            //     indicators:true
            // })
            // results.appendChild(carousel);
            // document.getElementsByClassName("carousel").carousel({
            //     fullWidth: true,
            //     indicators: true
            //   });
            var instance = M.Carousel.init(carousel, {                                
                indicators: true,
                duration: 400                                
            });                                                                                     

            

        }else if(data.icon != null){
            var image = document.createElement("img");
            image.classList.add("my-icon")
            image.src = data.icon.url;        
            card.appendChild(image);
        }
        content.appendChild(title);
        content.appendChild(description);    
        card.appendChild(content);
        results.appendChild(card);
    }

}

function showError(error) {
    results.innerText = "Error: " + error
}

