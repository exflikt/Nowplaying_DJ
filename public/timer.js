document.addEventListener("DOMContentLoaded", () => {
    //ノード単体に対する表示・非表示の設定を行う
    function display(elem, nweek, ntime){
        //ノードに設定されている曜日・時刻を取得する
        const data  = elem.dataset;
        const week  = data.week;
        const stime = ((data.sh|0) * 60 + (data.sm|0)) * 60;
        const etime = ((data.eh||24) * 60 + (data.em|0)) * 60 + (data.es|0);
        //表示時間判定
        if((!week || nweek === week) && (stime <= ntime && ntime <= etime)){
            //表示
            elem.style.display = "block";
        }else{
            //非表示
            elem.style.display = "none";
        }
    }

    const nodelist = document.querySelectorAll('.dtimer');
    //指定したノードリストを設定用関数(display)に送る
    (function displayAll(){
        const today = new Date();
        const nweek = ["日", "月", "火", "水", "木", "金", "土"][today.getDay()];
        const ntime = (today.getHours() * 60 + today.getMinutes()) * 60 + today.getSeconds();
        nodelist.forEach((elem) => display(elem, nweek, ntime));
        setTimeout(displayAll, 1000);
    }());
});
