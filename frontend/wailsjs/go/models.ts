export namespace main {
	
	export class changeDetail {
	    At: string;
	    AtShort: string;
	    Field: string;
	    From: string;
	    To: string;
	    ScanURL: string;
	
	    static createFrom(source: any = {}) {
	        return new changeDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.At = source["At"];
	        this.AtShort = source["AtShort"];
	        this.Field = source["Field"];
	        this.From = source["From"];
	        this.To = source["To"];
	        this.ScanURL = source["ScanURL"];
	    }
	}
	export class jobSummary {
	    Exists: boolean;
	    Min: string;
	    Avg: string;
	    Max: string;
	    Jitter: string;
	    Real: string;
	    AvgValue: number;
	
	    static createFrom(source: any = {}) {
	        return new jobSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Exists = source["Exists"];
	        this.Min = source["Min"];
	        this.Avg = source["Avg"];
	        this.Max = source["Max"];
	        this.Jitter = source["Jitter"];
	        this.Real = source["Real"];
	        this.AvgValue = source["AvgValue"];
	    }
	}
	export class payoutView {
	    output_index: number;
	    address: string;
	    amount_btc: number;
	    type: string;
	    IsWorker: boolean;
	    Percent: number;
	
	    static createFrom(source: any = {}) {
	        return new payoutView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.output_index = source["output_index"];
	        this.address = source["address"];
	        this.amount_btc = source["amount_btc"];
	        this.type = source["type"];
	        this.IsWorker = source["IsWorker"];
	        this.Percent = source["Percent"];
	    }
	}
	export class issueDetail {
	    Message: string;
	    Explanation: string;
	    Score: number;
	
	    static createFrom(source: any = {}) {
	        return new issueDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Message = source["Message"];
	        this.Explanation = source["Explanation"];
	        this.Score = source["Score"];
	    }
	}
	export class pingSummary {
	    Exists: boolean;
	    Min: string;
	    Avg: string;
	    Max: string;
	    Samples: number;
	    AvgValue: number;
	    Jitter: string;
	    RealPing: string;
	
	    static createFrom(source: any = {}) {
	        return new pingSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Exists = source["Exists"];
	        this.Min = source["Min"];
	        this.Avg = source["Avg"];
	        this.Max = source["Max"];
	        this.Samples = source["Samples"];
	        this.AvgValue = source["AvgValue"];
	        this.Jitter = source["Jitter"];
	        this.RealPing = source["RealPing"];
	    }
	}
	export class entryView {
	    Timestamp: string;
	    // Go type: time
	    TimestampRaw: any;
	    LogFile: string;
	    PoolName: string;
	    Host: string;
	    Port: number;
	    PortDisplay: string;
	    Ping: string;
	    PingSummaryPrimary: pingSummary;
	    PingSummaryTLS: pingSummary;
	    PingSort: number;
	    TotalPayout: number;
	    WorkerShare: number;
	    WorkerPercent: number;
	    PoolWallet: string;
	    PoolWalletDisp: string;
	    PoolWalletURL: string;
	    HasPoolWallet: boolean;
	    TLS: boolean;
	    ShowTLSPanel: boolean;
	    Issues: issueDetail[];
	    IssueSeverity: number;
	    Changes: changeDetail[];
	    HiddenChanges: number;
	    DisplayPayouts: payoutView[];
	    SplitCount: number;
	    HasData: boolean;
	    Connected: boolean;
	    Error: string;
	    PanelClass: string;
	    RewardNote: string;
	    RewardClass: string;
	    PingClass: string;
	    ScanURL: string;
	    HistoryURL: string;
	    LatestChanges: changeDetail[];
	    JobLatency: string;
	    JobLatencyClass: string;
	    JobWaitSummary: jobSummary;
	    JobWaitSort: number;
	
	    static createFrom(source: any = {}) {
	        return new entryView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Timestamp = source["Timestamp"];
	        this.TimestampRaw = this.convertValues(source["TimestampRaw"], null);
	        this.LogFile = source["LogFile"];
	        this.PoolName = source["PoolName"];
	        this.Host = source["Host"];
	        this.Port = source["Port"];
	        this.PortDisplay = source["PortDisplay"];
	        this.Ping = source["Ping"];
	        this.PingSummaryPrimary = this.convertValues(source["PingSummaryPrimary"], pingSummary);
	        this.PingSummaryTLS = this.convertValues(source["PingSummaryTLS"], pingSummary);
	        this.PingSort = source["PingSort"];
	        this.TotalPayout = source["TotalPayout"];
	        this.WorkerShare = source["WorkerShare"];
	        this.WorkerPercent = source["WorkerPercent"];
	        this.PoolWallet = source["PoolWallet"];
	        this.PoolWalletDisp = source["PoolWalletDisp"];
	        this.PoolWalletURL = source["PoolWalletURL"];
	        this.HasPoolWallet = source["HasPoolWallet"];
	        this.TLS = source["TLS"];
	        this.ShowTLSPanel = source["ShowTLSPanel"];
	        this.Issues = this.convertValues(source["Issues"], issueDetail);
	        this.IssueSeverity = source["IssueSeverity"];
	        this.Changes = this.convertValues(source["Changes"], changeDetail);
	        this.HiddenChanges = source["HiddenChanges"];
	        this.DisplayPayouts = this.convertValues(source["DisplayPayouts"], payoutView);
	        this.SplitCount = source["SplitCount"];
	        this.HasData = source["HasData"];
	        this.Connected = source["Connected"];
	        this.Error = source["Error"];
	        this.PanelClass = source["PanelClass"];
	        this.RewardNote = source["RewardNote"];
	        this.RewardClass = source["RewardClass"];
	        this.PingClass = source["PingClass"];
	        this.ScanURL = source["ScanURL"];
	        this.HistoryURL = source["HistoryURL"];
	        this.LatestChanges = this.convertValues(source["LatestChanges"], changeDetail);
	        this.JobLatency = source["JobLatency"];
	        this.JobLatencyClass = source["JobLatencyClass"];
	        this.JobWaitSummary = this.convertValues(source["JobWaitSummary"], jobSummary);
	        this.JobWaitSort = source["JobWaitSort"];
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
	export class hostView {
	    Host: string;
	    Latest?: entryView;
	
	    static createFrom(source: any = {}) {
	        return new hostView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Host = source["Host"];
	        this.Latest = this.convertValues(source["Latest"], entryView);
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
	export class hostEntry {
	    PoolName: string;
	    Host?: hostView;
	    LogFile: string;
	
	    static createFrom(source: any = {}) {
	        return new hostEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.PoolName = source["PoolName"];
	        this.Host = this.convertValues(source["Host"], hostView);
	        this.LogFile = source["LogFile"];
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
	export class dashboardView {
	    CleanEntries: hostEntry[];
	    IssueEntries: hostEntry[];
	    SortBy: string;
	    HostFilter: string;
	
	    static createFrom(source: any = {}) {
	        return new dashboardView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.CleanEntries = this.convertValues(source["CleanEntries"], hostEntry);
	        this.IssueEntries = this.convertValues(source["IssueEntries"], hostEntry);
	        this.SortBy = source["SortBy"];
	        this.HostFilter = source["HostFilter"];
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

