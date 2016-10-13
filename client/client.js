window.webrecording = {};

window.webrecording.Pipeline = class {
    constructor(next) {
        this.next = next;
    }

    consume(data) {
        this.next.consume(data);
    }

    start() {
        this.next.start();
    }

    stop() {
        this.next.stop();
    }
};

window.webrecording.Recorder = class extends webrecording.Pipeline {
    constructor(stream, next) {
        super(next);
        this.recorder = new MediaRecorder(stream);
        this.recorder.onstop = () => {
            super.stop();
            this.onstop();
        };
        this.recorder.onstart = () => {
            this.onstart();
        };
        this.recorder.ondataavailable = e => {
            this.next.consume(e.data);
        };
    }

    start() {
        super.start();
        this.recorder.start(1000);
    }

    stop() {
        this.recorder.stop();
    }

    onstart() {

    }

    onstop() {

    }

};

window.webrecording.Uploader = class extends webrecording.Pipeline {
    constructor() {
        super();
        this.guid = (() => {
            function s4() {
                return Math.floor((1 + Math.random()) * 0x10000)
                    .toString(16)
                    .substring(1);
            }

            return s4() + s4() + '-' + s4() + '-' + s4() + '-' +
                s4() + '-' + s4() + s4() + s4();
        })()
    }

    consume(data) {
        this.ws.send(new Blob([data]));
    }

    start() {
        this.ws = new WebSocket("ws://backend.localhost/");
        this.ws.onopen = () => this.ws.send(this.guid + '.webm');
    }

    stop() {
        this.ws.close();
    }
};

window.webrecording.default_pipeline = function (stream) {
    return new webrecording.Recorder(stream, new webrecording.Uploader());
};