window.webrecording = {};

window.webrecording.network = {buffered: [0]};

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
        let index = 0;
        this.recorder.ondataavailable = e => {
            super.consume({payload: e.data, index: index++});
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
        this.queue = [];
    }

    consume(data) {
        let last_elements = window.webrecording.network.buffered.slice(-5);
        if (last_elements.every((x) => x > 0)) {
            this.queue.push(data);
        } else {
            super.consume(data);
        }
    }

    stop() {
        for (const data of this.queue) {
            super.consume(data);
        }
        super.stop();
    }
};

window.webrecording.Uploader = class extends webrecording.Pipeline {
    constructor() {
        super();
        this.nwChecker = 0;
    }

    consume(data) {
        this.ws.send(new Blob([new Int32Array([data.index, data.payload.size]), data.payload]));
    }

    start() {
        this.ws = new WebSocket("ws://backend.localhost/");
        let guid = (() => {
            function s4() {
                return Math.floor((1 + Math.random()) * 0x10000)
                    .toString(16)
                    .substring(1);
            }

            return s4() + s4() + '-' + s4() + '-' + s4() + '-' +
                s4() + '-' + s4() + s4() + s4();
        })();
        this.ws.onopen = () => this.ws.send(guid + '.webm');
        this.nwChecker = setInterval(() => {
            let prev = webrecording.network.buffered.slice(-1)[0];
            webrecording.network.buffered.push(this.ws.bufferedAmount - prev);
        }, 200);
    }

    stop() {
        window.clearInterval(this.nwChecker);
        this.ws.send(new Int32Array([-1]).buffer);
        this.ws.close();
    }
};

window.webrecording.defaultPipeline = function (stream) {
    return new webrecording.Recorder(stream, new webrecording.BandwidthFilter(new webrecording.Uploader()));
};