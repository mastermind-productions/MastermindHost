import {useState} from 'react';
import {Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Stack} from '@mui/material';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import StopIcon from '@mui/icons-material/Stop';
import PowerSettingsNewIcon from '@mui/icons-material/PowerSettingsNew';
import RestartAltIcon from '@mui/icons-material/RestartAlt';
import DownloadIcon from '@mui/icons-material/Download';
import SystemUpdateAltIcon from '@mui/icons-material/SystemUpdateAlt';
import CancelIcon from '@mui/icons-material/Cancel';
import {backend, Host, isActive, run} from '../api';

type Confirm = {title: string; text: string; action: () => void};

export default function Actions({server, onChanged}: {server: backend.ServerInfo; onChanged: () => void}) {
    const [confirm, setConfirm] = useState<Confirm | null>(null);
    const s = server.status;
    const active = isActive(s);
    const working = s === 'installing' || s === 'updating';
    const transitional = ['stopping', 'shutting down', 'restarting'].includes(s);
    const go = (p: Promise<any>, msg?: string) => { run(p, msg).then(onChanged); };

    return (
        <Stack direction="row" spacing={1} sx={{flexWrap: 'wrap', rowGap: 1}}>
            {!server.installed && !server.script && !working && (
                <Button variant="contained" startIcon={<DownloadIcon/>} onClick={() => go(Host.Install(server.id, false), 'Install started')}>Install</Button>
            )}
            {working && (
                <Button color="warning" variant="outlined" startIcon={<CancelIcon/>} onClick={() => go(Host.CancelOperation(server.id))}>Cancel</Button>
            )}
            {server.installed && !active && !working && !transitional && (
                <Button variant="contained" startIcon={<PlayArrowIcon/>} onClick={() => go(Host.Start(server.id), 'Server starting')}>Start</Button>
            )}
            {active && (
                <>
                    <Button variant="outlined" startIcon={<PowerSettingsNewIcon/>} onClick={() => setConfirm({
                        title: 'Shutdown', text: server.script ? 'Ask every component to close (reverse start order), then force-stop any that remain?' : 'Save the world via RCON and exit gracefully?', action: () => go(Host.Shutdown(server.id), 'Server shut down'),
                    })}>Shutdown</Button>
                    <Button variant="outlined" startIcon={<RestartAltIcon/>} onClick={() => setConfirm({
                        title: 'Restart', text: server.script ? 'Close all components and run the start command again?' : 'Save, shut down and start the server again?', action: () => go(Host.Restart(server.id), 'Server restarted'),
                    })}>Restart</Button>
                </>
            )}
            {(active || transitional) && (
                <Button color="error" variant="outlined" startIcon={<StopIcon/>} onClick={() => setConfirm({
                    title: 'Force stop', text: server.script ? 'Close all components of this server now? Unsaved progress will be lost.' : 'Kill the server process immediately? Unsaved progress will be lost.', action: () => go(Host.Stop(server.id), 'Server stopped'),
                })}>Stop</Button>
            )}
            {server.installed && !server.script && !working && !transitional && (
                <Button variant={server.updateAvailable ? 'contained' : 'outlined'} color={server.updateAvailable ? 'warning' : 'primary'}
                        startIcon={<SystemUpdateAltIcon/>}
                        onClick={() => setConfirm({
                            title: 'Update', text: active ? 'The server will be shut down, updated and started again.' : 'Run SteamCMD update now?',
                            action: () => go(Host.Update(server.id, false), 'Update started'),
                        })}>{server.updateAvailable ? 'Update available' : 'Update'}</Button>
            )}
            <Dialog open={!!confirm} onClose={() => setConfirm(null)}>
                <DialogTitle>{confirm?.title}</DialogTitle>
                <DialogContent><DialogContentText>{confirm?.text}</DialogContentText></DialogContent>
                <DialogActions>
                    <Button onClick={() => setConfirm(null)}>Cancel</Button>
                    <Button variant="contained" onClick={() => { confirm?.action(); setConfirm(null); }}>Confirm</Button>
                </DialogActions>
            </Dialog>
        </Stack>
    );
}
