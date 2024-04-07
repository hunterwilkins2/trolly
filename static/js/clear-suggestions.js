const itemInput = document.getElementById('item');

itemInput.addEventListener("focusout", () => {
    const suggestions = document.getElementById('suggestions');
    setTimeout(() => {
        if (suggestions.firstChild) {
            suggestions.firstChild.style.display = 'none';
        }
    }, 100)
});
