/**
 * Created by lvass on 2016.10.14..
 */

webrecording = {};

webrecording.network = {buffered: [0]};

webrecording.Pipeline = class {
    constructor(next) {
        this.next = next;
    }

    consume(data) {
        this.next && this.next.consume(data);
    }

    start() {
        this.next && this.next.start();
    }

    stop() {
        this.next && this.next.stop();
    }
};

webrecording.BandwidthFilter = class extends webrecording.Pipeline {

    constructor(next) {
        super(next);
        this.queue = [];
    }

    consume(data) {
        let last_elements = webrecording.network.buffered.slice(-5);
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

webrecording.Transport = class {
    setup() {
    }

    close() {
    }

    send(data) {
    }

};

webrecording.Websocket = class extends webrecording.Transport {

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
        console.log(`Sending chunk ${data.index} of length ${data.payload.size}`);
        this.ws.send(new Blob([new Int32Array([data.index, data.payload.size]), data.payload]));
    }

    close() {
        clearInterval(this.nwChecker);
        this.ws.send(new Int32Array([-1]).buffer);
        this.ws.close();
    }
};

webrecording.Http = class extends webrecording.Transport {
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
        clearInterval(this.nwChecker);
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


webrecording.Uploader = class extends webrecording.Pipeline {

    constructor() {
        super();
        this.transport = new webrecording.Websocket();
        this.transport.fallback(() => {
            this.stop(true);
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
        clearInterval(this.nwChecker);
        if (!error) {
            this.guid = undefined;
            this.transport.close();
        }
    }
};

self.addEventListener('activate', function(event) {
    event.waitUntil(new Promise(resolve => {
        console.log('Worker activated');
        self.clients.claim();
        resolve();
    }));
});

self.addEventListener('message', function(event) {
    if (!self.uploader) self.uploader = new webrecording.BandwidthFilter(new webrecording.Uploader());
    if(event.data.command) {
        let command = event.data.command;
        console.log('Got command in worker: ', command);
        if(typeof self.uploader[command] === 'function') {
            self.uploader[command].call(self.uploader);
        }
    }
    else {
        console.log('Got data in worker of length: ', event.data.payload.size);
        self.uploader.consume(event.data);
    }

});
