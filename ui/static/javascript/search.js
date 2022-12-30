const searchField = document.getElementById("search")
const dataList = document.getElementById("items")

searchField.addEventListener("input", search)

async function search() {
    const query = searchField.value
    if (query === "") {
        return
    }

    const response = await fetch(`/search/${query}`)
    if (!response.ok) {
        return
    }

    const items = await response.json()
    let list = ''
    for(item of items) {
        list += `<option value="${item.name}" />\n`
    }

    dataList.innerHTML = list
}