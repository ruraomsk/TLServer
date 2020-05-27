'use strict';

let mainData;
let errors = [];

$(function () {
    $.ajax({
        url: window.location.href,
        type: 'POST',
        dataType: 'json',
        contentType: 'application/json',
        success: function (data) {
            // console.log(data);
            mainData = data.crossMap;
            // let crossMap = data.crossMap;

            let treeObject = [];
            let branchObject = {text: '', checked: false, id: -1, children: []};

            let mainCounter = 0;
            for (let region in mainData) {
                let areas = mainData[region];
                let mainBranch = Object.assign({}, branchObject);
                mainBranch.text = region;
                mainBranch.id = mainCounter++;
                mainBranch.children = [];
                for (let area in areas) {
                    let tflights = areas[area];
                    let secondaryBranch = {text: '', children: []};
                    secondaryBranch.text = area;
                    for (let tflight in tflights) {
                        let children = {text: tflights[tflight].description, checked: false};
                        secondaryBranch.children.push(children);
                    }
                    mainBranch.children.push(secondaryBranch);
                }
                treeObject.push(mainBranch);
            }

            let myTree = new TreeView(treeObject, {

                // always shows the checkboxes
                showAlwaysCheckBox: true,

                // is foldable?
                fold: true,

                // opens all nodes on init
                openAllFold: false

            });

            document.body.appendChild(myTree.root)
        },
        // data: JSON.stringify(toSend),
        error: function (request) {
            // if(request.responseText.message === 'Invalid login credentials') {
            //     $('#oldPasswordForm').append('<div style="color: red;" id="oldPasswordMsg"><h5>Неверный пароль</h5></div>');
            // }
            // if(request.responseText.message === 'Password contains invalid characters') {
            //     $('#newPasswordForm').append('<div style="color: red;" id="newPasswordMsg"><h5>Пароль содержит недопустимые символы</h5></div>');
            // }
            console.log(request.status + ' ' + request.responseText);
        }
    });

    $('#allButton').on('click', function () {
        $.ajax({
            url: window.location.href + '/checkAllCross',
            type: 'POST',
            dataType: 'json',
            contentType: 'application/json',
            success: function (data) {
                console.log(data);
                if(data.tf.length !== 0) errors = data.tf;
                // $('.trigger').show();
                // if ($('.panel').attr('style') !== 'display: block;') $('.trigger').trigger('click');
            },
            // data: JSON.stringify(toSend),
            error: function (request) {
                console.log(request.status + ' ' + request.responseText);
            }
        });
    });

    $('#selectedButton').on('click', function () {

        let toSend = {};
        let selected = [];
        // $('span > span > span[id*=""]').each(function () {
        //     console.log($(this)[0].description);
        // });

        $('span[check-value*="1"]').each(function () {
            if($(this)[0].innerText.includes('ДК')) selected.push($(this)[0].innerText);
        });

        for (let region in mainData) {
            let areas = mainData[region];
            for (let area in areas) {
                let tflights = areas[area];
                for (let tflight in tflights) {
                    let flag = false;
                    errors.forEach(error => {
                        if(error.description === tflights[tflight].description) flag = true;
                    });
                    if (flag) {
                        console.log(tflights[tflight].region.num + ' ' + tflights[tflight].area.num + ' ' + tflights[tflight].ID + ' ' + tflights[tflight].description);
                    }
                }
            }
        }

        // $.ajax({
        //     url: window.location.href + '/checkSelected',//make
        //     type: 'POST',
        //     dataType: 'json',
        //     contentType: 'application/json',
        //     success: function (data) {
        //         console.log(data);
        //         if(data.tf.length !== 0) errors = data.tf;
        //         // $('.trigger').show();
        //         // if ($('.panel').attr('style') !== 'display: block;') $('.trigger').trigger('click');
        //     },
        //     data: JSON.stringify(toSend),
        //     error: function (request) {
        //         console.log(request.status + ' ' + request.responseText);
        //     }
        // });
    })
});
