import {useEffect, useState} from 'react';
import {Button, Checkbox, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, FormControlLabel, Paper, Stack, Table, TableBody, TableCell, TableHead, TableRow, TextField, Typography} from '@mui/material';
import {backend, fmtBytes, Host, run} from '../../api';
import type {TabProps} from '../ServerPage';

export default function BackupsTab({server}: TabProps) {
    const [list, setList] = useState<backend.BackupInfo[]>([]);
    const [cfg, setCfg] = useState<backend.BackupSettings | null>(null);
    const [busy, setBusy] = useState(false);
    const [restore, setRestore] = useState<string | null>(null);

    const load = () => Host.ListBackups(server.id).then(setList);
    useEffect(() => {
        load();
        Host.GetServerRaw(server.id).then(s => setCfg(s.backup));
    }, [server.id]);

    const set = (p: Partial<backend.BackupSettings>) => cfg && setCfg(new backend.BackupSettings({...cfg, ...p}));

    return (
        <Stack spacing={2}>
            {cfg && (
                <Paper sx={{p: 2}}>
                    <Typography variant="subtitle1" sx={{fontWeight: 600, mb: 2}}>Backup settings</Typography>
                    <Stack spacing={1.5}>
                        <Stack direction="row" spacing={1}>
                            <TextField fullWidth label="Backup folder (empty = default)" value={cfg.dir} onChange={e => set({dir: e.target.value})}/>
                            <Button variant="outlined" onClick={async () => { const d = await run(Host.SelectFolder('Backup folder')); if (d) set({dir: d}); }}>Browse</Button>
                        </Stack>
                        <Stack direction="row" spacing={2} sx={{alignItems: 'center', flexWrap: 'wrap'}}>
                            <FormControlLabel label="Auto-remove old backups" control={<Checkbox checked={cfg.autoRemove} onChange={e => set({autoRemove: e.target.checked})}/>}/>
                            <TextField label="Keep newest" type="number" value={cfg.keep} onChange={e => set({keep: Number(e.target.value)})} sx={{width: 140}}/>
                            <FormControlLabel label="Skip dated world .ark copies" control={<Checkbox checked={cfg.skipDatedWorldArks} onChange={e => set({skipDatedWorldArks: e.target.checked})}/>}/>
                            <Button variant="contained" onClick={() => run(Host.SetBackupSettings(server.id, cfg), 'Backup settings saved').then(load)}>Save settings</Button>
                        </Stack>
                    </Stack>
                </Paper>
            )}
            <Paper>
                <Stack direction="row" spacing={1} sx={{p: 1.5, alignItems: 'center'}}>
                    <Typography variant="subtitle1" sx={{fontWeight: 600, flex: 1}}>Backups ({list.length})</Typography>
                    <Button onClick={() => run(Host.OpenBackupFolder(server.id))}>Open folder</Button>
                    <Button variant="contained" disabled={busy} onClick={async () => {
                        setBusy(true);
                        await run(Host.Backup(server.id), 'Backup created');
                        setBusy(false); load();
                    }}>{busy ? 'Backing up...' : 'Backup now'}</Button>
                </Stack>
                <Table size="small">
                    <TableHead><TableRow><TableCell>Name</TableCell><TableCell>Created</TableCell><TableCell>Size</TableCell><TableCell align="right">Actions</TableCell></TableRow></TableHead>
                    <TableBody>
                        {list.map(b => (
                            <TableRow key={b.name}>
                                <TableCell sx={{fontFamily: 'monospace'}}>{b.name}</TableCell>
                                <TableCell>{new Date(b.created).toLocaleString()}</TableCell>
                                <TableCell>{fmtBytes(b.size)}</TableCell>
                                <TableCell align="right">
                                    <Button onClick={() => setRestore(b.name)}>Restore</Button>
                                    <Button color="error" onClick={() => run(Host.DeleteBackup(server.id, b.name), 'Deleted').then(load)}>Delete</Button>
                                </TableCell>
                            </TableRow>
                        ))}
                        {!list.length && <TableRow><TableCell colSpan={4} sx={{color: 'text.secondary'}}>No backups yet.</TableCell></TableRow>}
                    </TableBody>
                </Table>
            </Paper>
            <Dialog open={!!restore} onClose={() => setRestore(null)}>
                <DialogTitle>Restore backup</DialogTitle>
                <DialogContent><DialogContentText>Overwrite the current saves and configs with {restore}? The server must be stopped.</DialogContentText></DialogContent>
                <DialogActions>
                    <Button onClick={() => setRestore(null)}>Cancel</Button>
                    <Button variant="contained" color="warning" onClick={() => { run(Host.RestoreBackup(server.id, restore!), 'Backup restored'); setRestore(null); }}>Restore</Button>
                </DialogActions>
            </Dialog>
        </Stack>
    );
}
