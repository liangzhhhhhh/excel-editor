export namespace dataparser {
	
	export class CellData {
	    v?: any;
	    s?: string;
	    t?: number;
	
	    static createFrom(source: any = {}) {
	        return new CellData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.v = source["v"];
	        this.s = source["s"];
	        this.t = source["t"];
	    }
	}
	export class ColumnStyle {
	    w?: number;
	
	    static createFrom(source: any = {}) {
	        return new ColumnStyle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.w = source["w"];
	    }
	}
	export class RowHeaderStyle {
	    width?: number;
	    hidden?: number;
	
	    static createFrom(source: any = {}) {
	        return new RowHeaderStyle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.width = source["width"];
	        this.hidden = source["hidden"];
	    }
	}
	export class RowStyle {
	    h?: number;
	    ia?: number;
	
	    static createFrom(source: any = {}) {
	        return new RowStyle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.h = source["h"];
	        this.ia = source["ia"];
	    }
	}
	export class Sheet {
	    name?: string;
	    id?: string;
	    rowCount?: number;
	    columnCount?: number;
	    cellData?: Record<number, any>;
	    columnData?: Record<number, ColumnStyle>;
	    rowData?: Record<number, RowStyle>;
	    zoomRatio?: number;
	    tabColor?: string;
	    scrollTop?: number;
	    scrollLeft?: number;
	    rowHeader?: RowHeaderStyle;
	
	    static createFrom(source: any = {}) {
	        return new Sheet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.id = source["id"];
	        this.rowCount = source["rowCount"];
	        this.columnCount = source["columnCount"];
	        this.cellData = source["cellData"];
	        this.columnData = this.convertValues(source["columnData"], ColumnStyle, true);
	        this.rowData = this.convertValues(source["rowData"], RowStyle, true);
	        this.zoomRatio = source["zoomRatio"];
	        this.tabColor = source["tabColor"];
	        this.scrollTop = source["scrollTop"];
	        this.scrollLeft = source["scrollLeft"];
	        this.rowHeader = this.convertValues(source["rowHeader"], RowHeaderStyle);
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
	export class Workbook {
	    id: string;
	    name?: string;
	    sheetOrder?: string[];
	    sheets?: Record<string, Sheet>;
	    styles?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new Workbook(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.sheetOrder = source["sheetOrder"];
	        this.sheets = this.convertValues(source["sheets"], Sheet, true);
	        this.styles = source["styles"];
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

export namespace types {
	
	export class CommonResponse {
	    status: number;
	    msg: string;
	    data: any;
	
	    static createFrom(source: any = {}) {
	        return new CommonResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.msg = source["msg"];
	        this.data = source["data"];
	    }
	}

}

