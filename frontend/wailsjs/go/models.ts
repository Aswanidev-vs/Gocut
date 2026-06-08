export namespace project {
	
	export class Asset {
	    id: string;
	    path: string;
	    type: string;
	    duration: number;
	    width: number;
	    height: number;
	    fps: number;
	    codec: string;
	    thumbnail: string;
	    waveform: number[];
	    fileSize: number;
	    // Go type: time
	    importedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Asset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.path = source["path"];
	        this.type = source["type"];
	        this.duration = source["duration"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.fps = source["fps"];
	        this.codec = source["codec"];
	        this.thumbnail = source["thumbnail"];
	        this.waveform = source["waveform"];
	        this.fileSize = source["fileSize"];
	        this.importedAt = this.convertValues(source["importedAt"], null);
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
	export class StickerProps {
	    x: number;
	    y: number;
	    width: number;
	    height: number;
	    rotation: number;
	    opacity: number;
	    flipH: boolean;
	    flipV: boolean;
	
	    static createFrom(source: any = {}) {
	        return new StickerProps(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.rotation = source["rotation"];
	        this.opacity = source["opacity"];
	        this.flipH = source["flipH"];
	        this.flipV = source["flipV"];
	    }
	}
	export class TextProps {
	    text: string;
	    fontFamily: string;
	    fontSize: number;
	    bold: boolean;
	    italic: boolean;
	    underline: boolean;
	    color: string;
	    strokeColor: string;
	    strokeWidth: number;
	    shadowColor: string;
	    shadowBlur: number;
	    shadowOffsetX: number;
	    shadowOffsetY: number;
	    bgColor: string;
	    bgPadding: number;
	    bgBorderRadius: number;
	    align: string;
	    letterSpacing: number;
	    lineHeight: number;
	
	    static createFrom(source: any = {}) {
	        return new TextProps(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.text = source["text"];
	        this.fontFamily = source["fontFamily"];
	        this.fontSize = source["fontSize"];
	        this.bold = source["bold"];
	        this.italic = source["italic"];
	        this.underline = source["underline"];
	        this.color = source["color"];
	        this.strokeColor = source["strokeColor"];
	        this.strokeWidth = source["strokeWidth"];
	        this.shadowColor = source["shadowColor"];
	        this.shadowBlur = source["shadowBlur"];
	        this.shadowOffsetX = source["shadowOffsetX"];
	        this.shadowOffsetY = source["shadowOffsetY"];
	        this.bgColor = source["bgColor"];
	        this.bgPadding = source["bgPadding"];
	        this.bgBorderRadius = source["bgBorderRadius"];
	        this.align = source["align"];
	        this.letterSpacing = source["letterSpacing"];
	        this.lineHeight = source["lineHeight"];
	    }
	}
	export class Transition {
	    type: string;
	    duration: number;
	
	    static createFrom(source: any = {}) {
	        return new Transition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.duration = source["duration"];
	    }
	}
	export class Keyframe {
	    id: string;
	    time: number;
	    property: string;
	    value: any;
	    easing: string;
	
	    static createFrom(source: any = {}) {
	        return new Keyframe(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.time = source["time"];
	        this.property = source["property"];
	        this.value = source["value"];
	        this.easing = source["easing"];
	    }
	}
	export class ColorGrade {
	    brightness: number;
	    contrast: number;
	    saturation: number;
	    hue: number;
	    sharpness: number;
	    vignette: number;
	    grain: number;
	    blur: number;
	    tint: number;
	    temp: number;
	    highlights: number;
	    shadows: number;
	    liftR: number;
	    liftG: number;
	    liftB: number;
	    gammaR: number;
	    gammaG: number;
	    gammaB: number;
	    gainR: number;
	    gainG: number;
	    gainB: number;
	    curves: string;
	    chromaKeyColor: string;
	    chromaKeySimilarity: number;
	    chromaKeyBlend: number;
	
	    static createFrom(source: any = {}) {
	        return new ColorGrade(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.brightness = source["brightness"];
	        this.contrast = source["contrast"];
	        this.saturation = source["saturation"];
	        this.hue = source["hue"];
	        this.sharpness = source["sharpness"];
	        this.vignette = source["vignette"];
	        this.grain = source["grain"];
	        this.blur = source["blur"];
	        this.tint = source["tint"];
	        this.temp = source["temp"];
	        this.highlights = source["highlights"];
	        this.shadows = source["shadows"];
	        this.liftR = source["liftR"];
	        this.liftG = source["liftG"];
	        this.liftB = source["liftB"];
	        this.gammaR = source["gammaR"];
	        this.gammaG = source["gammaG"];
	        this.gammaB = source["gammaB"];
	        this.gainR = source["gainR"];
	        this.gainG = source["gainG"];
	        this.gainB = source["gainB"];
	        this.curves = source["curves"];
	        this.chromaKeyColor = source["chromaKeyColor"];
	        this.chromaKeySimilarity = source["chromaKeySimilarity"];
	        this.chromaKeyBlend = source["chromaKeyBlend"];
	    }
	}
	export class Transform {
	    x: number;
	    y: number;
	    scaleX: number;
	    scaleY: number;
	    rotation: number;
	    flipH: boolean;
	    flipV: boolean;
	    cropX: number;
	    cropY: number;
	    cropW: number;
	    cropH: number;
	
	    static createFrom(source: any = {}) {
	        return new Transform(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	        this.scaleX = source["scaleX"];
	        this.scaleY = source["scaleY"];
	        this.rotation = source["rotation"];
	        this.flipH = source["flipH"];
	        this.flipV = source["flipV"];
	        this.cropX = source["cropX"];
	        this.cropY = source["cropY"];
	        this.cropW = source["cropW"];
	        this.cropH = source["cropH"];
	    }
	}
	export class Clip {
	    id: string;
	    assetId: string;
	    trackId: string;
	    startTime: number;
	    duration: number;
	    trimStart: number;
	    trimEnd: number;
	    speed: number;
	    reversed: boolean;
	    volume: number;
	    opacity: number;
	    transform: Transform;
	    color: ColorGrade;
	    keyframes: Keyframe[];
	    transition?: Transition;
	    textProps?: TextProps;
	    stickerProps?: StickerProps;
	
	    static createFrom(source: any = {}) {
	        return new Clip(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.assetId = source["assetId"];
	        this.trackId = source["trackId"];
	        this.startTime = source["startTime"];
	        this.duration = source["duration"];
	        this.trimStart = source["trimStart"];
	        this.trimEnd = source["trimEnd"];
	        this.speed = source["speed"];
	        this.reversed = source["reversed"];
	        this.volume = source["volume"];
	        this.opacity = source["opacity"];
	        this.transform = this.convertValues(source["transform"], Transform);
	        this.color = this.convertValues(source["color"], ColorGrade);
	        this.keyframes = this.convertValues(source["keyframes"], Keyframe);
	        this.transition = this.convertValues(source["transition"], Transition);
	        this.textProps = this.convertValues(source["textProps"], TextProps);
	        this.stickerProps = this.convertValues(source["stickerProps"], StickerProps);
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
	
	export class FileFilter {
	    name: string;
	    extensions: string[];
	
	    static createFrom(source: any = {}) {
	        return new FileFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.extensions = source["extensions"];
	    }
	}
	
	export class MediaInfo {
	    path: string;
	    duration: number;
	    width: number;
	    height: number;
	    fps: number;
	    codec: string;
	    audioCodec: string;
	    fileSize: number;
	
	    static createFrom(source: any = {}) {
	        return new MediaInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.duration = source["duration"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.fps = source["fps"];
	        this.codec = source["codec"];
	        this.audioCodec = source["audioCodec"];
	        this.fileSize = source["fileSize"];
	    }
	}
	export class ProjectSettings {
	    name: string;
	    aspectRatio: string;
	    resolution?: Resolution;
	    fps: number;
	    backgroundColor: string;
	    autoSave: boolean;
	    autoSaveIntervalSeconds: number;
	
	    static createFrom(source: any = {}) {
	        return new ProjectSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.aspectRatio = source["aspectRatio"];
	        this.resolution = this.convertValues(source["resolution"], Resolution);
	        this.fps = source["fps"];
	        this.backgroundColor = source["backgroundColor"];
	        this.autoSave = source["autoSave"];
	        this.autoSaveIntervalSeconds = source["autoSaveIntervalSeconds"];
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
	export class Track {
	    id: string;
	    type: string;
	    clips: Clip[];
	    muted: boolean;
	    locked: boolean;
	    volume: number;
	
	    static createFrom(source: any = {}) {
	        return new Track(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.clips = this.convertValues(source["clips"], Clip);
	        this.muted = source["muted"];
	        this.locked = source["locked"];
	        this.volume = source["volume"];
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
	export class Timeline {
	    tracks: Track[];
	    duration: number;
	
	    static createFrom(source: any = {}) {
	        return new Timeline(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tracks = this.convertValues(source["tracks"], Track);
	        this.duration = source["duration"];
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
	export class Resolution {
	    width: number;
	    height: number;
	
	    static createFrom(source: any = {}) {
	        return new Resolution(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.width = source["width"];
	        this.height = source["height"];
	    }
	}
	export class Project {
	    id: string;
	    name: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	    duration: number;
	    aspectRatio: string;
	    resolution: Resolution;
	    fps: number;
	    timeline: Timeline;
	    assets: Asset[];
	    settings: ProjectSettings;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.duration = source["duration"];
	        this.aspectRatio = source["aspectRatio"];
	        this.resolution = this.convertValues(source["resolution"], Resolution);
	        this.fps = source["fps"];
	        this.timeline = this.convertValues(source["timeline"], Timeline);
	        this.assets = this.convertValues(source["assets"], Asset);
	        this.settings = this.convertValues(source["settings"], ProjectSettings);
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
	
	export class RecentProject {
	    path: string;
	    name: string;
	    // Go type: time
	    updatedAt: any;
	    thumbnail?: string;
	
	    static createFrom(source: any = {}) {
	        return new RecentProject(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	        this.thumbnail = source["thumbnail"];
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
	export class RenderProgress {
	    jobId: string;
	    percent: number;
	    currentTime: number;
	    totalTime: number;
	    fps: number;
	    status: string;
	    error?: string;
	    outputPath?: string;
	
	    static createFrom(source: any = {}) {
	        return new RenderProgress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.jobId = source["jobId"];
	        this.percent = source["percent"];
	        this.currentTime = source["currentTime"];
	        this.totalTime = source["totalTime"];
	        this.fps = source["fps"];
	        this.status = source["status"];
	        this.error = source["error"];
	        this.outputPath = source["outputPath"];
	    }
	}
	export class RenderSettings {
	    jobId: string;
	    outputPath: string;
	    format: string;
	    codec: string;
	    width: number;
	    height: number;
	    fps: number;
	    bitrate: string;
	    audioBitrate: string;
	    crf: number;
	    preset: string;
	    startTime: number;
	    endTime: number;
	
	    static createFrom(source: any = {}) {
	        return new RenderSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.jobId = source["jobId"];
	        this.outputPath = source["outputPath"];
	        this.format = source["format"];
	        this.codec = source["codec"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.fps = source["fps"];
	        this.bitrate = source["bitrate"];
	        this.audioBitrate = source["audioBitrate"];
	        this.crf = source["crf"];
	        this.preset = source["preset"];
	        this.startTime = source["startTime"];
	        this.endTime = source["endTime"];
	    }
	}
	
	
	
	
	
	

}

