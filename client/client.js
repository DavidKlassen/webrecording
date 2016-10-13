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
            super.consume(e.data);
        };
    }

    start() {
        super.start();
        this.recorder.start(3000);
    }

    stop() {
        this.recorder.stop();
    }

    onstart() {

    }

    onstop() {

    }

};

window.webrecording.BandwidthFilter = class extends webrecording.Pipeline {

    constructor(next) {
        super(next);
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
        this.ws.send(new Blob([new Int32Array([0, data.size]), data]));
        this.ws.send(new Blob([data]));
    }

    start() {
        this.ws = new WebSocket("ws://backend.localhost/");
        this.ws.onopen = () => this.ws.send(this.guid + '.webm');
    }

    stop() {
        this.ws.send(new Int32Array([-1]).buffer);
        this.ws.close();
    }
};

window.webrecording.default_pipeline = function (stream) {
    return new webrecording.Recorder(stream, new webrecording.BandwidthFilter(new webrecording.Uploader()));
};