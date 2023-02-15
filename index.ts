import { ChildProcess, spawn} from "node:child_process";


export class TurnServer {

    private readonly process: ChildProcess; 

    constructor() {
        this.executable = this.setExecutable(options);
        this.executableOptions = this.setExecutableOptions(options);
        this.spawnOptions = { ...options.spawnOptions, stdio: 'pipe', detached: false, shell: false };
        this.invocationQueue = this.setInvocationQueue();

        this.process = spawn(this.executable, this.executableOptions, this.spawnOptions);
        this.streams = {
        stdin: this.process.stdin as Writable,
        stdout: this.process.stdout as Readable,
        stderr: this.process.stderr as Readable,
        };
    }

}
