document.addEventListener('DOMContentLoaded', function(event) {
    (function() {
        var widget_link, iframe, i, widget_links;
        widget_links = document.getElementsByClassName('mybooksrating-bookshelf');
        for (i = 0; i < widget_links.length; i++) {
            widget_link = widget_links[i];
            iframe = document.createElement('iframe');
            iframe.setAttribute('id', "bookshelf-" + i);
            iframe.setAttribute('class', "bookshelf");
            iframe.setAttribute('src', widget_link.href);
            //iframe.setAttribute('width', '300');
            iframe.setAttribute('height', '240');
            iframe.setAttribute('frameborder', '0');
            iframe.setAttribute('scrolling', 'no');

           // iframe.setAttribute('onload', `iframeLoaded("bookshelf-` + i + `")`)
            widget_link.parentNode.replaceChild(iframe, widget_link);
        }
    })();
})



window.addEventListener('message', function(e) {
    let message = e.data;
    bookIframes = document.getElementsByClassName('bookshelf');
    if (message.height) {
        for (i = 0; i < bookIframes.length; i++) {
            bookIframes[i].height = message.height + 'px';
            bookIframes[i].width = message.width + 'px';
        }
    }
    
} , false);

// function iframeLoaded(id) {
//     var iFrameID = document.getElementById(id);
//     if(iFrameID) {
//         // here you can make the height, I delete it first, then I make it again
//         iFrameID.height = "";
//         iFrameID.height = iFrameID.contentWindow.document.body.scrollHeight + "px";
//     }
// }
