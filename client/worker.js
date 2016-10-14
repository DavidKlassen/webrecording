/**
 * Created by lvass on 2016.10.13..
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
        clearInterval(this.nwChecker);
        this.ws.send(new Int32Array([-1]).buffer);
        this.ws.close();
    }
};


onmessage = function(e) {
    if(e.data.command) {
        let command = e.data.command;
        console.log('Got command in worker: ', command);
        if(typeof uploader[command] === 'function') {
            uploader[command].call(uploader);
        }
    }
    else {
        console.log('Got data in worker of length: ', e.data.payload.size);
        uploader.consume(e.data);
    }
}



