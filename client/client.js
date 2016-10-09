window.webrecording = {};
window.webrecording.Recorder = class {
    constructor(stream) {
        this.recorder = new MediaRecorder(stream);
        this.recorder.onstart = () => this.onstart();
        this.recorder.onstop = () => {
            this.ws.close();
            this.onstop();
        };
        this.recorder.ondataavailable = e => {
            this.ws.send(new Blob([e.data]));
            this.index++;
            this.chunks.push(e.data);
        };
    }

    start() {
        this.index = 0;
        this.chunks = [];
        this.ws = new WebSocket("ws://backend.localhost/");
        return this.recorder.start(1000);
    }

    stop() {
        this.recorder.stop();
    }
};
