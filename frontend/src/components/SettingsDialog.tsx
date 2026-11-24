import {useEffect, useState} from 'react';
import {Button, Dialog, DialogActions, DialogContent, DialogTitle, Stack, TextField, Typography} from '@mui/material';
import {backend, Host, run} from '../api';

export default function SettingsDialog({open, onClose}: {open: boolean; onClose: () => void}) {
    const [s, setS] = useState<backend.Settings>(new backend.Settings({serversRoot: '', backupsRoot: ''}));

    useEffect(() => { if (open) Host.GetSettings().then(setS); }, [open]);

    const pick = async (key: 'serversRoot' | 'backupsRoot') => {
        const d = await run(Host.SelectFolder('Select folder'));
        if (d) setS(new backend.Settings({...s, [key]: d}));
    };

    return (
        <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm">
            <DialogTitle>Settings</DialogTitle>
            <DialogContent>
                <Stack spacing={2} sx={{mt: 1}}>
                    {(['serversRoot', 'backupsRoot'] as const).map(k => (
                        <Stack key={k} direction="row" spacing={1}>
                            <TextField fullWidth label={k === 'serversRoot' ? 'Servers folder (new servers)' : 'Default backups folder'}
                                       value={s[k]} onChange={e => setS(new backend.Settings({...s, [k]: e.target.value}))}/>
                            <Button variant="outlined" onClick={() => pick(k)}>Browse</Button>
                        </Stack>
                    ))}
                    <Typography variant="body2" sx={{color: 'text.secondary'}}>
                        Existing servers keep their current folders. SteamCMD is downloaded automatically on first install.
                    </Typography>
                </Stack>
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>Cancel</Button>
                <Button variant="contained" onClick={async () => {
                    if (await run(Host.SetSettings(s).then(() => true), 'Settings saved')) onClose();
                }}>Save</Button>
            </DialogActions>
        </Dialog>
    );
}
