import { ChildProcess, spawn} from "node:child_process";
import { Readable, Writable } from "node:stream";


const EDGE_BIN = "./edge-turn"

export type StdioStreams = {
  stdin: Writable;
  stdout: Readable;
  stderr: Readable;
};


export class TurnServer {

    private process: ChildProcess | null; 
    private streams: StdioStreams | null;
    private timer  : NodeJS.Timer | null;
    private stop : boolean

    private readonly username : string; 
    private readonly password : string;
    private readonly port     : number;


    private exited : boolean
    private logfunc : (log : string) => (void);

    constructor(username: string, password: string, port: number,logfunc: (log:string)=>(void)) {
        this.username =username
        this.password = password
        this.port = port;
        this.exited = true
        this.process = null
        this.streams = null
        this.timer = null
        this.stop = false
        this.logfunc  = logfunc;
    }


    private async Spawn() {
        if (this.stop) {
            return
        }

        console.log(`spawning turn server on port ${this.port}`)
        this.process = spawn(EDGE_BIN, [
            `--port`,`${this.port}`,
            `--user`,`${this.username}`,
            `--password`,`${this.password}`
        ]);

        if (process == null) {
            console.log("fail to spawn edge-turn, maybe the executable is missing?")
            this.exited = true;
            return
        }
        this.exited = false;

        this.streams = {
            stdin: this.process.stdin as Writable,
            stdout: this.process.stdout as Readable,
            stderr: this.process.stderr as Readable,
        };

        this.streams.stdout.on("data",(data : Buffer) => {
            this.logfunc(data.toString());
        })
        this.streams.stderr.on("data",(data : Buffer) => {
            this.logfunc(data.toString());
        })

        this.process?.on("spawn",() => {
            console.log(`spawned turn server`)
        })
        this.process?.on("error",(err) => {
            console.log(`server error ${err.message}`);
        })
        this.process?.on("close",(data) => {
            console.log(`turn server closed with code : ${data}`)
            this.exited = true
        })
        this.process?.on("exit",(exit) => {
            console.log(`turn server exited with code : ${exit}`)
            this.exited = true
        })          
    }

    public async Start() {
        this.stop = false;
        this.timer = setInterval(async ()=>{
            if (this.exited) {
                await this.Spawn()
            }
        },2000)
    }

    public async Stop() {
        this.stop = true;
        if (this.timer != null) {
            clearInterval(this.timer)
        }

        this.process?.kill()
    }
}





