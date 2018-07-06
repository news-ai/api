window.Intercom("boot", {
  app_id: "ur8dbk9e"
});

function alturaMaxima() {
    var altura = $(window).height();
    $(".full-screen").css('min-height', altura);
}

$(document).ready(function() {
    alturaMaxima();
    $(window).bind('resize', alturaMaxima);
});

$.urlParam = function(name){
    var results = new RegExp('[\?&]' + name + '=([^&#]*)').exec(window.location.href);
    if (results==null){
       return null;
    }
    else{
       return results[1] || 0;
    }
}

function insertParam(key, value) {
    key = encodeURI(key); value = encodeURI(value);
    var kvp = document.location.search.substr(1).split('&');
    var i=kvp.length; var x; while(i--) {
        x = kvp[i].split('=');

        if (x[0]==key){
            x[1] = value;
            kvp[i] = x.join('=');
            break;
        }
    }

    if(i<0) {kvp[kvp.length] = [key,value].join('=');}

    //this will reload the page, it's likely better to store this until finished
    document.location.search = kvp.join('&'); 
}

if (navigator.userAgent.match(/IEMobile\/10\.0/)) {
    var msViewportStyle = document.createElement('style')
    msViewportStyle.appendChild(
        document.createTextNode(
            '@-ms-viewport{width:auto!important}'
        )
    )
    document.querySelector('head').appendChild(msViewportStyle)
}

function changePricingClass() {
    if (document.getElementById("pricingClass").className === "annually") {
        // Change label color & the button
        document.getElementById("pricingClass").className = "monthly";
        document.getElementById("monthlyLabel").className = "active";
        document.getElementById("annuallyLabel").className = "";

        // Update prices
        document.getElementById("personalPrice").innerHTML = "18.99";
        document.getElementById("consultantPrice").innerHTML = "34.99";
        document.getElementById("businessPrice").innerHTML = "41.99";
        document.getElementById("ultimatePrice").innerHTML = "52.99";

        $('.duration').attr('value', 'monthly');
    } else {
        // Change label color & the button
        document.getElementById("pricingClass").className = "annually";
        document.getElementById("monthlyLabel").className = "";
        document.getElementById("annuallyLabel").className = "active";

        // Update prices
        document.getElementById("personalPrice").innerHTML = "15.99";
        document.getElementById("consultantPrice").innerHTML = "28.99";
        document.getElementById("businessPrice").innerHTML = "34.99";
        document.getElementById("ultimatePrice").innerHTML = "43.99";

        $('.duration').attr('value', 'annually');
    }
}

// Check if there's an invitation code
var inviteCode = $.urlParam('invite');
if (inviteCode) {
    document.getElementById("mce-INVITE").value = inviteCode;
}