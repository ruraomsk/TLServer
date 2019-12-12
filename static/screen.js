"use strict";

function check() {

	var account = {
		login: $("#login").val(),
		password: $("#password").val()
	};
	$.ajax({
		type: 'POST',
		url: 'http://192.168.1.220:8082/login',
		data: JSON.stringify(account),
		dataType: 'json',
		success: function (data) {
			//                document.cookie = data.token;
			sessionStorage.setItem("token", data.token);
//			console.log(data.message + " " + data.token);
//			testToken();
            alert(window.location.href);
            goToPage();
		},
		error: function (request, errorMsg) {
			console.log(errorMsg);
		}
	});
};

function goToPage() {
    location.href = 'http://192.168.1.220:8082/' + $("#login").val();
}