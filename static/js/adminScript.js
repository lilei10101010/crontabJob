$(function(){
    /**
     *  当文件上传的input改变内容的时候上传文件
     */
    $('.ajaxUploadImage').change(function(){
        var _this = $(this);
        LoadingStart("uploading...");
        var data = new FormData();
        data.append('myfile',$(this).get(0).files[0]);//获取当前的input的files,后台接收的时候name是"myfile"
        //data.append('isUploadImage','1');//可自己增加的选项

        // ajax上传部分
        $.ajax({
            url:'/UploadFile',//你的上传方法
            type:'POST',
            data:data,
            cache: false,//选项必填
            contentType: false,//选项必填
            processData: false,//选项必填
            success:function(data){
                console.log(data);
                layer.msg(data.msg);
                $("input[name='fileValue']").val(data.url);//返回的数值
                LoadingEnd();
            },
            error:function(){
                LoadingEnd();
                commAlert("Upload failed, please try again");
            }
        });

    });

});

function LoadingStart(LoadText='Loading...'){
    var loadingHidder = null;
    loadingHidder = "<div class='commLoading' style='background: rgba(0,0,0,0.1); width: 100%;height: 100%;position: fixed;top: 0;left: 0;z-index: 100;'><div style='background: #fff;position: absolute;top: 50%;left: 50%; transform: translate(-50%,-50%);-webkit-transform: translate(-50%,-50%);padding: 30px 50px 20px;border-radius: 5px;line-height: 20px;box-sizing: border-box;text-align: center;'><img src='/static/image/loading.gif' style='width: 35px;display: block;margin: 0 auto 15px;' /><span>" + LoadText + "<span></div></div>";
    $("body").css("overflow","hidden").append(loadingHidder);
}

function LoadingEnd(){
    $("body").css("overflow","");
    $(".commLoading").remove();
}
function commAlert(ALERTINNERHTML){
    var alertHTML = "";
    alertHTML += "<div class='commAlertMask'><div class='commMaskBox'><div class='commMaskClose'><i style='font-size:22px;font-style:normal'>x</i></div><div className='commAlertCont'>" + ALERTINNERHTML + "</div></div></div>";

    $("body").append(alertHTML);

    $(".commAlertMask .commMaskClose").click(function(){
        $(".commAlertMask").remove();
    });
}
function commToastText(ToastText){
    var toastHTML = "";
    toastHTML += "<div class='commToastMask' style='background: rgba(0,0,0,0.5); width: 100%;height: 100%;position: fixed;top: 0;left: 0;z-index: 100;'><div style='background: #fff;position: absolute;top: 50%;left: 50%; transform: translate(-50%,-50%);-webkit-transform: translate(-50%,-50%);padding: 30px 50px;border-radius: 5px;line-height: 20px;box-sizing: border-box;text-align: center;'>" + ToastText + "</div></div>";
    $("body").append(toastHTML);

    setTimeout(function(){
        $(".commToastMask").remove();
    },2000);

}
function commToast(TOAST_HTML){
    var toastBlackHtml = "";
    toastBlackHtml += "<div class='commToastBlack' style='background: rgba(255,255,255,0.5);width: 100%;height: 100%;position: fixed;top: 0;left: 0;z-index: 100;'><div style='background: #000;color: #fff;position: absolute;top: 50%;left: 50%;transform: translate(-50%,-50%);-webkit-transform:translate(-50%,-50%); max-width: 800px;max-height: 90%;overflow: auto;padding: 20px 50px;border-radius: 5px;line-height: 20px;box-sizing: border-box;-webkit-box-sizing: border-box;'>" + TOAST_HTML + "</div></div>";
    $("body").append(toastBlackHtml);

    setTimeout(function(){
        $(".commToastBlack").remove();
    },2000);
}