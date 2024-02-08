document.addEventListener("DOMContentLoaded", () => {
    const scrollpos = sessionStorage.getItem("scrollpos");
    if (scrollpos) window.scrollTo(0, scrollpos);
});

window.onbeforeunload = () => {
    sessionStorage.setItem("scrollpos", window.scrollY);
};

function hotReload() {
    let socket = new WebSocket("ws://localhost:4000/hot-reload");

    socket.onopen = () => {
        console.debug("Hot reload active");
    };

    socket.onclose = () => {
        console.debug("Hot reload websocket closed. Trying to connect...");
        let counter = 0;
        let interval = setInterval(() => {
            if (counter === 20) {
                clearInterval(interval);
                console.debug(
                    "Hot reload connection could not be made. Reload page to establish again"
                );
            }

            fetch("http://localhost:4000/hot-reload/ready")
                .then((response) => {
                    if (response.status === 200) {
                        location.replace(location.href);
                    }
                })
                .catch((err) => { });
            counter++;
        }, 250);
    };

    socket.onerror = (error) => {
        console.debug("Socket error: ", error);
        socket.close();
    };
}

hotReload();