export namespace backend {
	
	export class A2SPlayer {
	    index: number;
	    name: string;
	    score: number;
	    duration: number;
	
	    static createFrom(source: any = {}) {
	        return new A2SPlayer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.name = source["name"];
	        this.score = source["score"];
	        this.duration = source["duration"];
	    }
	}
	export class A2SResult {
	    port: number;
	    name: string;
	    map: string;
	    folder: string;
	    game: string;
	    appId: number;
	    players: number;
	    maxPlayers: number;
	    bots: number;
	    serverType: string;
	    os: string;
	    passworded: boolean;
	    vac: boolean;
	    version: string;
	    pingMs: number;
	    playerList: A2SPlayer[];
	
	    static createFrom(source: any = {}) {
	        return new A2SResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.port = source["port"];
	        this.name = source["name"];
	        this.map = source["map"];
	        this.folder = source["folder"];
	        this.game = source["game"];
	        this.appId = source["appId"];
	        this.players = source["players"];
	        this.maxPlayers = source["maxPlayers"];
	        this.bots = source["bots"];
	        this.serverType = source["serverType"];
	        this.os = source["os"];
	        this.passworded = source["passworded"];
	        this.vac = source["vac"];
	        this.version = source["version"];
	        this.pingMs = source["pingMs"];
	        this.playerList = this.convertValues(source["playerList"], A2SPlayer);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AdvancedIniChange {
	    file: string;
	    section: string;
	    key: string;
	    value: string;
	    remove: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AdvancedIniChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file = source["file"];
	        this.section = source["section"];
	        this.key = source["key"];
	        this.value = source["value"];
	        this.remove = source["remove"];
	    }
	}
	export class AdvancedIniEntry {
	    file: string;
	    fileName: string;
	    section: string;
	    key: string;
	    value: string;
	    set: boolean;
	    group: string;
	    desc: string;
	    kind: string;
	
	    static createFrom(source: any = {}) {
	        return new AdvancedIniEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file = source["file"];
	        this.fileName = source["fileName"];
	        this.section = source["section"];
	        this.key = source["key"];
	        this.value = source["value"];
	        this.set = source["set"];
	        this.group = source["group"];
	        this.desc = source["desc"];
	        this.kind = source["kind"];
	    }
	}
	export class AffinityInfo {
	    cores: number;
	    saved: number[];
	    current: number[];
	    running: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AffinityInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cores = source["cores"];
	        this.saved = source["saved"];
	        this.current = source["current"];
	        this.running = source["running"];
	    }
	}
	export class KV {
	    key: string;
	    value: string;
	    hasValue: boolean;
	
	    static createFrom(source: any = {}) {
	        return new KV(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.value = source["value"];
	        this.hasValue = source["hasValue"];
	    }
	}
	export class ArkCmd {
	    map: string;
	    options: KV[];
	    flags: KV[];
	
	    static createFrom(source: any = {}) {
	        return new ArkCmd(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.map = source["map"];
	        this.options = this.convertValues(source["options"], KV);
	        this.flags = this.convertValues(source["flags"], KV);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BackupInfo {
	    name: string;
	    path: string;
	    size: number;
	    created: number;
	
	    static createFrom(source: any = {}) {
	        return new BackupInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.size = source["size"];
	        this.created = source["created"];
	    }
	}
	export class BackupSettings {
	    dir: string;
	    autoRemove: boolean;
	    keep: number;
	    skipDatedWorldArks: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BackupSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dir = source["dir"];
	        this.autoRemove = source["autoRemove"];
	        this.keep = source["keep"];
	        this.skipDatedWorldArks = source["skipDatedWorldArks"];
	    }
	}
	export class ConfigFileDef {
	    key: string;
	    name: string;
	    relPath: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigFileDef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.name = source["name"];
	        this.relPath = source["relPath"];
	        this.description = source["description"];
	    }
	}
	export class CurseForgeMod {
	    id: string;
	    name: string;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new CurseForgeMod(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.url = source["url"];
	    }
	}
	export class DiscordHook {
	    url: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DiscordHook(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.enabled = source["enabled"];
	    }
	}
	export class Event {
	    // Go type: time
	    time: any;
	    action: string;
	    result: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new Event(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time = this.convertValues(source["time"], null);
	        this.action = source["action"];
	        this.result = source["result"];
	        this.message = source["message"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class GameDef {
	    id: string;
	    name: string;
	    appId: string;
	    exe: string;
	    processName: string;
	    savesRel: string;
	    logRel: string;
	    header: string;
	    configFiles: ConfigFileDef[];
	    enabled: boolean;
	    kind: string;
	    defaultPort: number;
	    defaultRcon: number;
	    maps: string[];
	    group: string;
	    defaultDir: string;
	
	    static createFrom(source: any = {}) {
	        return new GameDef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.appId = source["appId"];
	        this.exe = source["exe"];
	        this.processName = source["processName"];
	        this.savesRel = source["savesRel"];
	        this.logRel = source["logRel"];
	        this.header = source["header"];
	        this.configFiles = this.convertValues(source["configFiles"], ConfigFileDef);
	        this.enabled = source["enabled"];
	        this.kind = source["kind"];
	        this.defaultPort = source["defaultPort"];
	        this.defaultRcon = source["defaultRcon"];
	        this.maps = source["maps"];
	        this.group = source["group"];
	        this.defaultDir = source["defaultDir"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class IniChange {
	    section: string;
	    key: string;
	    value: string;
	    remove: boolean;
	
	    static createFrom(source: any = {}) {
	        return new IniChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.section = source["section"];
	        this.key = source["key"];
	        this.value = source["value"];
	        this.remove = source["remove"];
	    }
	}
	
	export class LogFile {
	    name: string;
	    size: number;
	    modified: number;
	
	    static createFrom(source: any = {}) {
	        return new LogFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.size = source["size"];
	        this.modified = source["modified"];
	    }
	}
	export class MetricPoint {
	    t: number;
	    cpu: number;
	    ram: number;
	
	    static createFrom(source: any = {}) {
	        return new MetricPoint(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.t = source["t"];
	        this.cpu = source["cpu"];
	        this.ram = source["ram"];
	    }
	}
	export class ModInfo {
	    id: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new ModInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class Player {
	    index: string;
	    name: string;
	    id: string;
	
	    static createFrom(source: any = {}) {
	        return new Player(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.name = source["name"];
	        this.id = source["id"];
	    }
	}
	export class PortEntry {
	    label: string;
	    proto: string;
	    port: number;
	    bound: boolean;
	    pid: number;
	    process: string;
	    ours: boolean;
	    conflicts: string;
	
	    static createFrom(source: any = {}) {
	        return new PortEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.proto = source["proto"];
	        this.port = source["port"];
	        this.bound = source["bound"];
	        this.pid = source["pid"];
	        this.process = source["process"];
	        this.ours = source["ours"];
	        this.conflicts = source["conflicts"];
	    }
	}
	export class Schedule {
	    id: string;
	    enabled: boolean;
	    action: string;
	    mode: string;
	    time: string;
	    days: number[];
	    intervalMin: number;
	    arg: string;
	    warnMinutes: number;
	    // Go type: time
	    lastRun: any;
	
	    static createFrom(source: any = {}) {
	        return new Schedule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.enabled = source["enabled"];
	        this.action = source["action"];
	        this.mode = source["mode"];
	        this.time = source["time"];
	        this.days = source["days"];
	        this.intervalMin = source["intervalMin"];
	        this.arg = source["arg"];
	        this.warnMinutes = source["warnMinutes"];
	        this.lastRun = this.convertValues(source["lastRun"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ScriptComponent {
	    name: string;
	    exe: string;
	    port: number;
	    running: boolean;
	    listening: boolean;
	    pid: number;
	
	    static createFrom(source: any = {}) {
	        return new ScriptComponent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.exe = source["exe"];
	        this.port = source["port"];
	        this.running = source["running"];
	        this.listening = source["listening"];
	        this.pid = source["pid"];
	    }
	}
	export class ScriptProc {
	    name: string;
	    exe: string;
	    port: number;
	
	    static createFrom(source: any = {}) {
	        return new ScriptProc(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.exe = source["exe"];
	        this.port = source["port"];
	    }
	}
	export class ScriptSpec {
	    workDir: string;
	    startCmd: string;
	    stopCmd: string;
	    procs: ScriptProc[];
	
	    static createFrom(source: any = {}) {
	        return new ScriptSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workDir = source["workDir"];
	        this.startCmd = source["startCmd"];
	        this.stopCmd = source["stopCmd"];
	        this.procs = this.convertValues(source["procs"], ScriptProc);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WindowRect {
	    X: number;
	    Y: number;
	    W: number;
	    H: number;
	    Max: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WindowRect(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.X = source["X"];
	        this.Y = source["Y"];
	        this.W = source["W"];
	        this.H = source["H"];
	        this.Max = source["Max"];
	    }
	}
	export class Server {
	    id: string;
	    name: string;
	    game: string;
	    dir: string;
	    cmd: string;
	    // Go type: time
	    createdAt: any;
	    forceRunning: boolean;
	    betaBranch: string;
	    betaPassword: string;
	    discord: DiscordHook;
	    backup: BackupSettings;
	    modNames: Record<string, string>;
	    events: Event[];
	    schedules: Schedule[];
	    affinity: number[];
	    windowRect?: WindowRect;
	    script?: ScriptSpec;
	
	    static createFrom(source: any = {}) {
	        return new Server(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.game = source["game"];
	        this.dir = source["dir"];
	        this.cmd = source["cmd"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.forceRunning = source["forceRunning"];
	        this.betaBranch = source["betaBranch"];
	        this.betaPassword = source["betaPassword"];
	        this.discord = this.convertValues(source["discord"], DiscordHook);
	        this.backup = this.convertValues(source["backup"], BackupSettings);
	        this.modNames = source["modNames"];
	        this.events = this.convertValues(source["events"], Event);
	        this.schedules = this.convertValues(source["schedules"], Schedule);
	        this.affinity = source["affinity"];
	        this.windowRect = this.convertValues(source["windowRect"], WindowRect);
	        this.script = this.convertValues(source["script"], ScriptSpec);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ServerInfo {
	    id: string;
	    name: string;
	    game: string;
	    gameName: string;
	    dir: string;
	    createdAt: string;
	    status: string;
	    installed: boolean;
	    map: string;
	    sessionName: string;
	    port: number;
	    rconPort: number;
	    rconEnabled: boolean;
	    maxPlayers: number;
	    playerCount: number;
	    buildId: string;
	    latestBuildId: string;
	    updateAvailable: boolean;
	    cpu: number;
	    ram: number;
	    startTime: number;
	    forceRunning: boolean;
	    progress: number;
	    modCount: number;
	    clusterId: string;
	    hasPassword: boolean;
	    betaBranch: string;
	    header: string;
	    group: string;
	    script: boolean;
	    compTotal: number;
	    compUp: number;
	
	    static createFrom(source: any = {}) {
	        return new ServerInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.game = source["game"];
	        this.gameName = source["gameName"];
	        this.dir = source["dir"];
	        this.createdAt = source["createdAt"];
	        this.status = source["status"];
	        this.installed = source["installed"];
	        this.map = source["map"];
	        this.sessionName = source["sessionName"];
	        this.port = source["port"];
	        this.rconPort = source["rconPort"];
	        this.rconEnabled = source["rconEnabled"];
	        this.maxPlayers = source["maxPlayers"];
	        this.playerCount = source["playerCount"];
	        this.buildId = source["buildId"];
	        this.latestBuildId = source["latestBuildId"];
	        this.updateAvailable = source["updateAvailable"];
	        this.cpu = source["cpu"];
	        this.ram = source["ram"];
	        this.startTime = source["startTime"];
	        this.forceRunning = source["forceRunning"];
	        this.progress = source["progress"];
	        this.modCount = source["modCount"];
	        this.clusterId = source["clusterId"];
	        this.hasPassword = source["hasPassword"];
	        this.betaBranch = source["betaBranch"];
	        this.header = source["header"];
	        this.group = source["group"];
	        this.script = source["script"];
	        this.compTotal = source["compTotal"];
	        this.compUp = source["compUp"];
	    }
	}
	export class ServerPorts {
	    id: string;
	    name: string;
	    gameName: string;
	    running: boolean;
	    firewall: boolean;
	    ports: PortEntry[];
	
	    static createFrom(source: any = {}) {
	        return new ServerPorts(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.gameName = source["gameName"];
	        this.running = source["running"];
	        this.firewall = source["firewall"];
	        this.ports = this.convertValues(source["ports"], PortEntry);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Settings {
	    serversRoot: string;
	    backupsRoot: string;
	    window?: WindowRect;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serversRoot = source["serversRoot"];
	        this.backupsRoot = source["backupsRoot"];
	        this.window = this.convertValues(source["window"], WindowRect);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

