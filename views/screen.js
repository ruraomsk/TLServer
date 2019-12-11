"use strict";

function check(){
//    alert("Allo");
//        $("#submit").click(function(){
        var account = {login:$("#login").val(), password:$("#password").val()};
//        var jjson = JSON.stringify(account);
//            alert($("#login").val() + " : " + $("#password").val());
//        alert(JSON.stringify(account));
//            $.post("http://192.168.1.220:8082/login",
//                   jjson,
//                   function(response){
//                   alert("PRIEM");
////                     alert(login + "//" + message + "//" + status + "//" + token);
//                   }
//                   );
//                   alert(JSON.stringify(account));
//        });
        $.ajax({
            type: 'POST',
            url: 'http://192.168.1.220:8082/login',
            data: JSON.stringify(account),
            dataType: 'json',
            success: function(data) {
                document.cookie = data.token;
                console.log(data.message + " " + data.token);
            },
            error: function(request, errorMsg){
                console.log(errorMsg);
            }
        });

//        var tokenObj = {token:token};

        $.ajax({
            type: 'POST',
//            beforeSend: function(request){ alert(document.cookie);
//                 request.setRequestHeader("Authorization", "Bearer" + " " + document.cookie);
//            },
            headers: {"Authorization" : "Bearer" + " " + document.cookie},
            url: 'http://192.168.1.220:8082/testtoken',
//            data: JSON.stringify(tokenObj),
//            dataType: 'json',
            success: function(data) {
                console.log(data.message + " -2222222- " + document.cookie);
            },
            error: function(request, errorMsg){
                console.log(errorMsg + " -2222222- ");
            }
        });

};