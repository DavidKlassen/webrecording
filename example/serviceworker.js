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

webrecording.Uploader = class extends webrecording.Pipeline {
    constructor() {
        super();
        this.nwChecker = 0;
    }

    consume(data) {
        let timeout = this.ws.readyState ? 0 : 500;
        setTimeout(() => {
            console.log(`Sending chunk ${data.index} of length ${data.payload.size}`);
            this.ws.send(new Blob([new Int32Array([data.index, data.payload.size]), data.payload]));
        }, timeout);
    }

    start() {
        return new Promise((resolve, reject) => {
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
            this.ws.onopen = () => {
                this.ws.send(guid + '.webm');
                resolve();
            }
            this.nwChecker = setInterval(() => {
                let prev = webrecording.network.buffered.slice(-1)[0];
                webrecording.network.buffered.push(this.ws.bufferedAmount - prev);
            }, 200);
        })
    }

    stop() {
        clearInterval(this.nwChecker);
        this.ws.send(new Int32Array([-1]).buffer);
        this.ws.close();
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
    if (!self.uploader) self.uploader = new webrecording.Uploader();
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
