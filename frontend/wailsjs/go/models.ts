export namespace app {
	
	export class State {
	    rows: number;
	    cols: number;
	    cursor: editor.Cursor;
	    dirty: boolean;
	    filename: string;
	    lines: editor.RowToken[][];
	
	    static createFrom(source: any = {}) {
	        return new State(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rows = source["rows"];
	        this.cols = source["cols"];
	        this.cursor = this.convertValues(source["cursor"], editor.Cursor);
	        this.dirty = source["dirty"];
	        this.filename = source["filename"];
	        this.lines = this.convertValues(source["lines"], editor.RowToken);
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
	export class FileActionResult {
	    state: State;
	    path: string;
	    cancelled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FileActionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = this.convertValues(source["state"], State);
	        this.path = source["path"];
	        this.cancelled = source["cancelled"];
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

export namespace editor {
	
	export class Cursor {
	    row: number;
	    col: number;
	
	    static createFrom(source: any = {}) {
	        return new Cursor(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.row = source["row"];
	        this.col = source["col"];
	    }
	}
	export class RowToken {
	    col: number;
	    width: number;
	    text: string;
	    color?: string;
	    bgColor?: string;
	    bold?: boolean;
	    italic?: boolean;
	    underline?: boolean;
	    inverse?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RowToken(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.col = source["col"];
	        this.width = source["width"];
	        this.text = source["text"];
	        this.color = source["color"];
	        this.bgColor = source["bgColor"];
	        this.bold = source["bold"];
	        this.italic = source["italic"];
	        this.underline = source["underline"];
	        this.inverse = source["inverse"];
	    }
	}

}

