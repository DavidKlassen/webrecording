window.webrecording = {};
window.webrecording.Recorder = class {
    constructor(stream) {
        this.recorder = new MediaRecorder(stream);
        this.recorder.onstart = () => this.onstart();
        this.recorder.onstop = () => {
            this.ws.send(new Int32Array([-1]).buffer);
            this.ws.close();
            this.onstop();
        };
        let index = 0;
        this.recorder.ondataavailable = e => {
            this.ws.send(new Blob([new Int32Array([index++, e.data.size]), e.data]));
            this.index++;
            this.chunks.push(e.data);
        };
    }

    start() {
        this.index = 0;
        this.chunks = [];
        this.ws = new WebSocket("ws://backend.localhost/");
        this.ws.onopen = () => this.ws.send('test.webm');

        return this.recorder.start(1000);
    }

    stop() {
        this.recorder.stop();
    }
};
