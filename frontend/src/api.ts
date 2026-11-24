import * as Host from '../wailsjs/go/backend/Host';
import {backend} from '../wailsjs/go/models';
import {EventsOn} from '../wailsjs/runtime/runtime';

export {Host, backend, EventsOn};

type Notifier = (msg: string, severity: 'success' | 'error' | 'info') => void;
let notifier: Notifier = () => {};
export const setNotifier = (n: Notifier) => { notifier = n; };
export const notify: Notifier = (m, s) => notifier(m, s);

export async function run<T>(p: Promise<T>, success?: string): Promise<T | undefined> {
    try {
        const r = await p;
        if (success) notify(success, 'success');
        return r;
    } catch (e: any) {
        notify(String(e?.message ?? e), 'error');
        return undefined;
    }
}

export const fmtBytes = (n: number) => {
    if (!n) return '0 B';
    const u = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.min(u.length - 1, Math.floor(Math.log(n) / Math.log(1024)));
    return `${(n / Math.pow(1024, i)).toFixed(i ? 1 : 0)} ${u[i]}`;
};

export const fmtUptime = (startMs: number) => {
    if (!startMs) return '-';
    let s = Math.max(0, Math.floor((Date.now() - startMs) / 1000));
    const d = Math.floor(s / 86400); s %= 86400;
    const h = Math.floor(s / 3600); s %= 3600;
    const m = Math.floor(s / 60);
    return d ? `${d}d ${h}h ${m}m` : h ? `${h}h ${m}m` : `${m}m ${s % 60}s`;
};

export const statusColor = (s: string): 'success' | 'warning' | 'error' | 'info' | 'default' => {
    switch (s) {
        case 'online': return 'success';
        case 'running': case 'starting': case 'restarting': case 'stopping': case 'shutting down': return 'warning';
        case 'installing': case 'updating': return 'info';
        case 'not found': return 'error';
        default: return 'default';
    }
};

export const isActive = (s: string) => ['online', 'running', 'starting'].includes(s);
export const isBusy = (s: string) => ['installing', 'updating', 'stopping', 'shutting down', 'restarting'].includes(s);
