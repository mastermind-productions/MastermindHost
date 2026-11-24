import {useEffect, useRef, useState} from 'react';
import {Box, Button, FormControlLabel, MenuItem, Stack, Switch, TextField} from '@mui/material';
import {backend, fmtBytes, Host, run} from '../../api';
import type {TabProps} from '../ServerPage';

export default function LogsTab({server}: TabProps) {
    const [files, setFiles] = useState<backend.LogFile[]>([]);
    const [name, setName] = useState('');
    const [text, setText] = useState('');
    const [follow, setFollow] = useState(true);
    const [filter, setFilter] = useState('');
    const ref = useRef<HTMLPreElement>(null);

    useEffect(() => {
        Host.ListLogs(server.id).then(f => { setFiles(f); if (f.length && !name) setName(f.find(x => x.name === 'ShooterGame.log')?.name ?? f[0].name); });
    }, [server.id]);

    const load = () => name && Host.ReadLog(server.id, name, 512 * 1024).then(setText).catch(() => {});
    useEffect(() => {
        load();
        if (!follow) return;
        const t = setInterval(load, 3000);
        return () => clearInterval(t);
    }, [name, follow]);
    useEffect(() => { if (follow) ref.current?.scrollTo(0, ref.current.scrollHeight); }, [text, follow]);

    const shown = filter ? text.split('\n').filter(l => l.toLowerCase().includes(filter.toLowerCase())).join('\n') : text;

    return (
        <Stack spacing={1.5} sx={{height: '100%'}}>
            <Stack direction="row" spacing={1} sx={{alignItems: 'center'}}>
                <TextField select label="Log file" value={name} onChange={e => setName(e.target.value)} sx={{minWidth: 320}}>
                    {files.map(f => <MenuItem key={f.name} value={f.name}>{f.name} ({fmtBytes(f.size)})</MenuItem>)}
                </TextField>
                <TextField label="Filter" value={filter} onChange={e => setFilter(e.target.value)} sx={{flex: 1}}/>
                <FormControlLabel label="Follow" control={<Switch checked={follow} onChange={e => setFollow(e.target.checked)}/>}/>
                <Button onClick={load}>Reload</Button>
                <Button onClick={() => run(Host.OpenFolder(server.id, 'ShooterGame\\Saved\\Logs'))}>Open folder</Button>
            </Stack>
            <Box component="pre" ref={ref} sx={{flex: 1, minHeight: 400, m: 0, p: 1, overflow: 'auto', bgcolor: '#000', borderRadius: 1, fontSize: 12, fontFamily: 'Consolas, monospace', color: '#8fe3a8', whiteSpace: 'pre-wrap'}}>
                {files.length ? shown : 'No log files yet - start the server first.'}
            </Box>
        </Stack>
    );
}
