export namespace terminal {
	
	export class SessionInfo {
	    id: string;
	    worktreeId: string;
	    cwd: string;
	    state: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    lastActive: any;
	
	    static createFrom(source: any = {}) {
	        return new SessionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.worktreeId = source["worktreeId"];
	        this.cwd = source["cwd"];
	        this.state = source["state"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.lastActive = this.convertValues(source["lastActive"], null);
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

export namespace worktree {
	
	export class CommitInfo {
	    hash: string;
	    message: string;
	    author: string;
	    // Go type: time
	    time: any;
	
	    static createFrom(source: any = {}) {
	        return new CommitInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hash = source["hash"];
	        this.message = source["message"];
	        this.author = source["author"];
	        this.time = this.convertValues(source["time"], null);
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
	export class CreateOptions {
	    name: string;
	    branch: string;
	    baseBranch: string;
	    createNew: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CreateOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.branch = source["branch"];
	        this.baseBranch = source["baseBranch"];
	        this.createNew = source["createNew"];
	    }
	}
	export class FileStat {
	    path: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new FileStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.status = source["status"];
	    }
	}
	export class Worktree {
	    id: string;
	    name: string;
	    path: string;
	    branch: string;
	    head: string;
	    isMain: boolean;
	    hasChanges: boolean;
	    unpushed: number;
	    // Go type: time
	    lastActivity: any;
	
	    static createFrom(source: any = {}) {
	        return new Worktree(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.branch = source["branch"];
	        this.head = source["head"];
	        this.isMain = source["isMain"];
	        this.hasChanges = source["hasChanges"];
	        this.unpushed = source["unpushed"];
	        this.lastActivity = this.convertValues(source["lastActivity"], null);
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
	export class WorktreeStatus {
	    worktreeId: string;
	    branch: string;
	    head: string;
	    ahead: number;
	    behind: number;
	    staged: FileStat[];
	    unstaged: FileStat[];
	    untracked: string[];
	    lastCommit?: CommitInfo;
	
	    static createFrom(source: any = {}) {
	        return new WorktreeStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.worktreeId = source["worktreeId"];
	        this.branch = source["branch"];
	        this.head = source["head"];
	        this.ahead = source["ahead"];
	        this.behind = source["behind"];
	        this.staged = this.convertValues(source["staged"], FileStat);
	        this.unstaged = this.convertValues(source["unstaged"], FileStat);
	        this.untracked = source["untracked"];
	        this.lastCommit = this.convertValues(source["lastCommit"], CommitInfo);
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

