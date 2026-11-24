import {useEffect, useRef, useState} from 'react';
import {Box, Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, FormControlLabel, Paper, Stack, Switch, TextField, Typography} from '@mui/material';
import FolderOpenIcon from '@mui/icons-material/FolderOpen';
import {EventsOn, fmtBytes, fmtUptime, Host, run} from '../../api';
import {Field} from '../ServerCard';
import type {TabProps} from '../ServerPage';

export default function OverviewTab({server: s, onChanged}: TabProps) {
    const [name, setName] = useState(s.name);
    const [log, setLog] = useState<string[]>([]);
    const [del, setDel] = useState<'remove' | 'delete' | null>(null);
    const logRef = useRef<HTMLPreElement>(null);

    useEffect(() => {
        Host.GetInstallLog(s.id).then(setLog);
        return EventsOn('gs-install-log', (id: string, line: string) => {
            if (id === s.id) setLog(l => [...l.slice(-499), line]);
        });
    }, [s.id]);
    useEffect(() => { logRef.current?.scrollTo(0, logRef.current.scrollHeight); }, [log]);

    return (
        <Stack spacing={2}>
            <Paper sx={{p: 2}}>
                <Box sx={{display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(150px, 1fr))', gap: 2}}>
                    <Field label="Session name" value={s.sessionName || '-'}/>
                    <Field label="Map" value={s.map || '-'}/>
                    <Field label="Players" value={`${s.playerCount} / ${s.maxPlayers}`}/>
                    <Field label="Game port (UDP)" value={String(s.port)}/>
                    <Field label="RCON port" value={`${s.rconPort}${s.rconEnabled ? '' : ' (disabled)'}`}/>
                    <Field label="Password" value={s.hasPassword ? 'Yes' : 'No'}/>
                    <Field label="Cluster ID" value={s.clusterId || '-'}/>
                    <Field label="Installed build" value={s.buildId || '-'}/>
                    <Field label="Latest build" value={s.latestBuildId || '-'}/>
                    <Field label="Branch" value={s.betaBranch || 'public'}/>
                    <Field label="CPU" value={s.startTime ? `${s.cpu.toFixed(1)}%` : '-'}/>
                    <Field label="RAM" value={s.startTime ? fmtBytes(s.ram) : '-'}/>
                    <Field label="Uptime" value={fmtUptime(s.startTime)}/>
                    <Field label="Created" value={new Date(s.createdAt).toLocaleString()}/>
                </Box>
            </Paper>
            <Paper sx={{p: 2}}>
                <Stack spacing={2}>
                    <Stack direction="row" spacing={1}>
                        <TextField label="Server name (also renames the folder)" value={name} onChange={e => setName(e.target.value)} sx={{flex: 1}}/>
                        <Button variant="contained" disabled={!name.trim() || name === s.name}
                                onClick={() => run(Host.SetName(s.id, name), 'Renamed').then(onChanged)}>Rename</Button>
                    </Stack>
                    <FormControlLabel control={<Switch checked={s.forceRunning} onChange={e => run(Host.SetForceRunning(s.id, e.target.checked)).then(onChanged)}/>}
                                      label="Force running - automatically restart the server if it crashes"/>
                    <Stack direction="row" spacing={1} sx={{flexWrap: 'wrap', rowGap: 1}}>
                        <Button variant="outlined" startIcon={<FolderOpenIcon/>} onClick={() => run(Host.OpenFolder(s.id, ''))}>Server folder</Button>
                        <Button variant="outlined" startIcon={<FolderOpenIcon/>} onClick={() => run(Host.OpenFolder(s.id, 'ShooterGame\\Saved'))}>Saved folder</Button>
                        <Button variant="outlined" disabled={!s.installed} onClick={() => run(Host.Install(s.id, true), 'Validate started').then(onChanged)}>Validate files</Button>
                        <Box sx={{flex: 1}}/>
                        <Button color="error" variant="outlined" onClick={() => setDel('remove')}>Remove from list</Button>
                        <Button color="error" variant="contained" onClick={() => setDel('delete')}>Delete server</Button>
                    </Stack>
                </Stack>
            </Paper>
            <Paper sx={{p: 2}}>
                <Typography variant="subtitle2" sx={{mb: 1}}>SteamCMD output</Typography>
                <Box component="pre" ref={logRef} sx={{m: 0, p: 1, height: 240, overflow: 'auto', bgcolor: '#000', borderRadius: 1, fontSize: 12, fontFamily: 'Consolas, monospace', color: '#8fe3a8'}}>
                    {log.length ? log.join('\n') : 'No install activity yet.'}
                </Box>
            </Paper>
            <Dialog open={!!del} onClose={() => setDel(null)}>
                <DialogTitle>{del === 'delete' ? 'Delete server' : 'Remove server'}</DialogTitle>
                <DialogContent>
                    <DialogContentText>
                        {del === 'delete' ? `Permanently delete "${s.name}" and ALL files in ${s.dir}? This cannot be undone.`
                            : `Remove "${s.name}" from MastermindHost? Files on disk are kept.`}
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button onClick={() => setDel(null)}>Cancel</Button>
                    <Button color="error" variant="contained" onClick={async () => {
                        await run(del === 'delete' ? Host.Delete(s.id) : Host.Remove(s.id), del === 'delete' ? 'Server deleted' : 'Server removed');
                        setDel(null); onChanged();
                    }}>{del === 'delete' ? 'Delete' : 'Remove'}</Button>
                </DialogActions>
            </Dialog>
        </Stack>
    );
}
