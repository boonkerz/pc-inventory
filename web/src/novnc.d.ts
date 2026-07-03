// noVNC liefert keine TypeScript-Typen mit – minimale Ambient-Deklaration.
declare module "@novnc/novnc" {
  export default class RFB extends EventTarget {
    constructor(target: HTMLElement, url: string, options?: { credentials?: { password?: string } });
    scaleViewport: boolean;
    clipViewport: boolean;
    background: string;
    disconnect(): void;
    sendCredentials(creds: { password?: string }): void;
    sendCtrlAltDel(): void;
  }
}
