"use strict";
//
//function check() {
//    var login = document.getElementById("login").value;
//    var password = document.getElementById("password").value;
//
//    if(login==="" || password===""){
////        alert("Пожалуйста, заполните поля авторизации.");
//    } else {
//        var account = {login:login, password:password};
//        var test = document.createElement("P");
//        test.innerText = JSON.stringify(account);
//        document.body.appendChild(test);
//        alert(JSON.stringify(account));
//    }
//}
//T<FYSQ D HJN <KZNM
    $(function(){
        $("#submit").click(function(){
        var account = {login:$("#login").val(), password:$("#password").val()};
//            alert($("#login").val() + " : " + $("#password").val());
//        alert(JSON.stringify(account));
            $.post("/login",
                   JSON.stringify(account),
                   function(){
                       alert("Success!");
                   });
        });
    });