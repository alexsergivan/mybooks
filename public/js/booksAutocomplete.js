const autoCompleteJS = new autoComplete({
    name: "Books",
    data: {                              // Data src [Array, Function, Async] | (REQUIRED)
        src: async () => {
            // User search query
            document
                .querySelector("#status").innerHTML = "<div class=\"lds-ellipsis\"><div></div><div></div><div></div><div></div></div>";
            const query = document.querySelector("#autoComplete").value;
            // Fetch External Data Source
            const source = await fetch(`/api/books/search?q=${query}`);
            // Format data into JSON
            const data = await source.json();

            document
                .querySelector("#status").innerHTML = "";
            // Return Fetched data
            return data;
        },
        key: ["Title"],
        cache: false
    },

    // sort: (a, b) => {                    // Sort rendered results ascendingly | (Optional)
    //     if (a.match < b.match) return -1;
    //     if (a.match > b.match) return 1;
    //     return 0;
    // },
    trigger: {
        event: ["input"]
    },
    placeHolder: "Book search",     // Place Holder text                 | (Optional)
    selector: "#autoComplete",           // Input field selector              | (Optional)
    threshold: 3,                      // Min. Chars length to start Engine | (Optional)
    debounce: 500,                       // Post duration for engine to start | (Optional)
    searchEngine: function (query, record) {
        return record
    },              // Search Engine type/mode           | (Optional)
    resultsList: {                       // Rendered results list object      | (Optional)
        container: source => {
           // source.setAttribute("class", "w-full inline-block z-50")

        },
        destination: "#status",
        element: "div",
        idName: "books_list",
        className: "books_list w-full inline-block z-50",
        render: true,
        maxResults: 20,
    },


    resultItem: {
        highlight: {
            render: true,                    // Highlight matching results        | (Optional)
        },
        content: (data, element) => {
            element.setAttribute("class", "p-4 hover:bg-purple-800 inline-block cursor-pointer w-full grid grid-cols-6 gap-4 bg-white shadow mb-2 rounded-md bg-purple-700 border-2 border-purple-400")
            data.value["plainTitle"] = data.value.Title
            data.value[data.key] = data.match

            // Modify Results Item Content
            element.innerHTML = generateBookItem(data.value);
        },
        element: "div",
    },

    onSelection: feedback => {             // Action script onSelection event | (Optional)
        document.querySelector(".selection").innerHTML = `<h4 class="text-sm font-semibold mb-3 mt-5 purple-100">Selected book:</h4><div class="p-4 hover:bg-purple-800 inline-block cursor-pointer w-full grid grid-cols-6 gap-4 bg-purple-700 shadow mb-2 rounded-md border-2 border-purple-400">` + generateBookItem(feedback.selection.value, feedback.selection.value.plainTitle) + `</div>`;
        // Replace Input value with the selected value
        document.querySelector("#autoComplete").value = feedback.selection.value.plainTitle;
        document.querySelector("#bookID").value = feedback.selection.value.GoogleID;
    }
});


function generateBookItem(data, title) {
    var authors = ""

    if (data.Authors) {
        authors = data.Authors
    }
    finalTitle = data.Title;
    if (title) {
        finalTitle = title
    }
    return  `<div class="col-span-1"><img class="w-full h-auto" src="${data.Thumbnail}"/></div> 
             <div class="col-span-5 text-left content-center flex flex-wrap">
               <div>
                <div class="text-base text-purple-50 font-bold tracking-wide"> ${finalTitle}</div>
                <div class="text-purple-100 text-sm">${data.Subtitle}</div>
                <div class="text-purple-200 text-sm uppercase">${authors}</div>
               </div>
                
                
             </div> `
}