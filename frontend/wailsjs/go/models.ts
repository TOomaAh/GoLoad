export namespace app {
	
	export class AppInfo {
	    Version: string;
	    Name: string;
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Version = source["Version"];
	        this.Name = source["Name"];
	    }
	}

}

export namespace download {
	
	export class Download {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    url: string;
	    filename: string;
	    location: string;
	    status: string;
	    progress: number;
	    speed: number;
	    size: number;
	    downloaded: number;
	    // Go type: time
	    start_time?: any;
	    // Go type: time
	    end_time?: any;
	    eta?: string;
	    attempt: number;
	    max_attempts: number;
	    error_message?: string;
	
	    static createFrom(source: any = {}) {
	        return new Download(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.url = source["url"];
	        this.filename = source["filename"];
	        this.location = source["location"];
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.speed = source["speed"];
	        this.size = source["size"];
	        this.downloaded = source["downloaded"];
	        this.start_time = this.convertValues(source["start_time"], null);
	        this.end_time = this.convertValues(source["end_time"], null);
	        this.eta = source["eta"];
	        this.attempt = source["attempt"];
	        this.max_attempts = source["max_attempts"];
	        this.error_message = source["error_message"];
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

export namespace settings {
	
	export class Settings {
	    ID: number;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	    // Go type: gorm
	    DeletedAt: any;
	    darkMode: boolean;
	    downloadPath: string;
	    maxDownloads: number;
	    maxSpeed: number;
	    autoStartDownloads: boolean;
	    retryAfterFailure: boolean;
	    notificationEndDownload: boolean;
	    notificationErrorDownload: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
	        this.DeletedAt = this.convertValues(source["DeletedAt"], null);
	        this.darkMode = source["darkMode"];
	        this.downloadPath = source["downloadPath"];
	        this.maxDownloads = source["maxDownloads"];
	        this.maxSpeed = source["maxSpeed"];
	        this.autoStartDownloads = source["autoStartDownloads"];
	        this.retryAfterFailure = source["retryAfterFailure"];
	        this.notificationEndDownload = source["notificationEndDownload"];
	        this.notificationErrorDownload = source["notificationErrorDownload"];
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

