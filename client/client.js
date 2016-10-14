window.webrecording = {};

window.webrecording.network = {buffered: [0]};

window.webrecording.restrictions = {
    video: {
        width: {min: 320, max: 320},
        height: {min: 240, max: 240},
        frameRate: {min: 10, max: 10}
    }
};

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
        this.queue = [];
        super.stop();
    }
};

window.webrecording.Transport = class {
    setup() {
    }

    close() {
    }

    send(data) {
    }

};

window.webrecording.Websocket = class extends webrecording.Transport {

    fallback(func) {
        this._fallback = func;
    }

    setup(guid) {
        if (!WebSocket) {
            this._fallback();
            return;
        }
        this.ws = new WebSocket("ws://backend.localhost/");
        this.ws.onopen = () => this.ws.send(guid + '.webm');
        this.ws.onerror = () => this._fallback();
        this.nwChecker = setInterval(() => {
            let prev = webrecording.network.buffered.slice(-1)[0];
            webrecording.network.buffered.push(this.ws.bufferedAmount - prev);
        }, 200);
    }

    send(data) {
        this.ws.send(new Blob([new Int32Array([data.index, data.payload.size]), data.payload]));
    }

    close() {
        window.clearInterval(this.nwChecker);
        this.ws.send(new Int32Array([-1]).buffer);
        this.ws.close();
    }
};

window.webrecording.Http = class extends webrecording.Transport {
    setup(guid) {
        this.guid = guid;
    }

    send(data) {
        let headers = new Headers();
        headers.append('X-Name', this.guid + '.webm');
        fetch("http://backend.localhost/chunk",
            {
                method: "POST",
                headers: headers,
                body: new Blob([new Int32Array([data.index, data.payload.size]), data.payload],
                    {type: 'application/octet-binary'})
            });
    }

    close() {
        window.clearInterval(this.nwChecker);
        let headers = new Headers();
        headers.append('X-Name', this.guid + '.webm');
        let dummy = new Uint8Array([253, 0, 128, 1]);
        fetch("http://backend.localhost/chunk",
            {
                method: "POST",
                headers: headers,
                body: new Blob([new Int32Array([-1, dummy.length]), dummy], {type: 'application/octet-binary'})
            })
    }
};


window.webrecording.Uploader = class extends webrecording.Pipeline {

    constructor() {
        super();
        this.nwChecker = 0;
        this.transport = new webrecording.Websocket();
        this.transport.fallback(() => {
            this.transport.close(true);
            this.transport = new webrecording.Http();
            this.transport.setup(this.guid);
        });
    }

    consume(data) {
        this.transport.send(data);
    }

    start() {
        this.guid = this.guid || (() => {
                function s4() {
                    return Math.floor((1 + Math.random()) * 0x10000)
                        .toString(16)
                        .substring(1);
                }

                return s4() + s4() + '-' + s4() + '-' + s4() + '-' +
                    s4() + '-' + s4() + s4() + s4();
            })();
        this.transport.setup(this.guid);
    }

    stop(error = false) {
        window.clearInterval(this.nwChecker);
        if (!error) {
            this.guid = undefined;
            this.transport.close();
        }
    }
};

window.webrecording.defaultPipeline = function (stream) {
    return new webrecording.Recorder(stream, new webrecording.BandwidthFilter(new webrecording.Uploader()));
};