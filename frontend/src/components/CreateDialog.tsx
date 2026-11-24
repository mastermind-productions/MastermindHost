import {useEffect, useState} from 'react';
import {Button, Dialog, DialogActions, DialogContent, DialogTitle, MenuItem, Stack, Tab, Tabs, TextField, Typography} from '@mui/material';
import {backend, Host, run} from '../api';

export default function CreateDialog({open, onClose, onCreated}: {open: boolean; onClose: () => void; onCreated: (id: string) => void}) {
    const [tab, setTab] = useState(0);
    const [games, setGames] = useState<backend.GameDef[]>([]);
    const [game, setGame] = useState('arksa');
    const [name, setName] = useState('');
    const [dir, setDir] = useState('');
    const [root, setRoot] = useState('');
    const [busy, setBusy] = useState(false);

    const gd = games.find(g => g.id === game);
    const isScript = gd?.kind === 'script';
    const pickGame = (id: string) => {
        setGame(id);
        const g = games.find(x => x.id === id);
        if (g?.kind === 'script') { setTab(1); setDir(g.defaultDir); setName(g.name); }
        else if (isScript) { setDir(''); setName(''); }
    };

    useEffect(() => {
        if (!open) return;
        setName(''); setDir(''); setTab(0);
        Host.GetGames().then(g => { setGames(g); if (g.length) setGame(g[0].id); });
        Host.GetSettings().then(s => setRoot(s.serversRoot));
    }, [open]);

    const submit = async () => {
        setBusy(true);
        const id = await run(tab === 0 ? Host.Create(name, game) : Host.Import(name, dir, game), tab === 0 ? 'Server created' : 'Server imported');
        setBusy(false);
        if (id) onCreated(id);
    };

    const folder = name.trim().replace(/[<>:"/\\|?*\x00-\x1f]/g, '_').replace(/[. ]+$/, '');

    return (
        <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm">
            <DialogTitle>Add Game Server</DialogTitle>
            <DialogContent>
                <Tabs value={tab} onChange={(_, v) => setTab(v)} sx={{mb: 2}}>
                    <Tab label="Create new" disabled={isScript}/>
                    <Tab label="Import existing"/>
                </Tabs>
                <Stack spacing={2}>
                    <TextField select label="Game" value={game} onChange={e => pickGame(e.target.value)}>
                        {games.map(g => <MenuItem key={g.id} value={g.id}>{g.name}</MenuItem>)}
                    </TextField>
                    <TextField label="Server name" value={name} autoFocus onChange={e => setName(e.target.value)}
                               onKeyDown={e => { if (e.key === 'Enter' && name.trim() && tab === 0) submit(); }}/>
                    {tab === 0 ? (
                        <Typography variant="body2" sx={{color: 'text.secondary', wordBreak: 'break-all'}}>
                            Folder: {root}\{folder || '...'}
                        </Typography>
                    ) : (
                        <Stack direction="row" spacing={1}>
                            <TextField label={isScript ? 'Server folder (contains the start script / executables)' : 'Existing server folder'} value={dir} onChange={e => setDir(e.target.value)} fullWidth/>
                            <Button variant="outlined" onClick={async () => { const d = await run(Host.SelectFolder('Select server folder')); if (d) setDir(d); }}>Browse</Button>
                        </Stack>
                    )}
                </Stack>
            </DialogContent>
            <DialogActions>
                <Button onClick={onClose}>Cancel</Button>
                <Button variant="contained" disabled={busy || (tab === 0 ? !name.trim() : !dir.trim())} onClick={submit}>
                    {tab === 0 ? 'Create' : 'Import'}
                </Button>
            </DialogActions>
        </Dialog>
    );
}
