import {useEffect, useState} from 'react';
import {Box, Button, Chip, Paper, Stack, TextField, Typography} from '@mui/material';
import {backend, Host, run} from '../../api';
import type {TabProps} from '../ServerPage';

const LISTS: [string, string][] = [
    ['adminList', 'Admins'],
    ['exclusiveList', 'Whitelist'],
    ['exclusiveListNoCheckList', 'Join no-check list'],
    ['banList', 'Ban list'],
];

export default function ListsTab({server}: TabProps) {
    const [files, setFiles] = useState<backend.ConfigFileDef[]>([]);
    useEffect(() => { Host.GetConfigFiles(server.id).then(f => setFiles(f ?? [])); }, [server.id]);
    return (
        <Box sx={{display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(420px, 1fr))', gap: 2}}>
            {LISTS.map(([key, title]) => {
                const f = files.find(x => x.key === key);
                return f && <ListEditor key={key} id={server.id} listKey={key} title={title} desc={`${f.name} - ${f.description}`}/>;
            })}
        </Box>
    );
}

function ListEditor({id, listKey, title, desc}: {id: string; listKey: string; title: string; desc: string}) {
    const [items, setItems] = useState<string[]>([]);
    const [val, setVal] = useState('');
    const load = () => Host.ReadList(id, listKey).then(setItems);
    useEffect(() => { load(); }, [id, listKey]);
    const save = (list: string[]) => run(Host.WriteList(id, listKey, list), `${title} saved`).then(load);
    const add = () => {
        const add = val.split(/[\s,]+/).filter(Boolean);
        if (add.length) save([...items, ...add]);
        setVal('');
    };
    return (
        <Paper sx={{p: 2}}>
            <Typography variant="subtitle1" sx={{fontWeight: 600}}>{title}</Typography>
            <Typography variant="caption" sx={{color: 'text.secondary'}}>{desc}</Typography>
            <Stack direction="row" spacing={1} sx={{my: 1.5}}>
                <TextField fullWidth placeholder="EOS / player ID" value={val} onChange={e => setVal(e.target.value)} onKeyDown={e => e.key === 'Enter' && add()}/>
                <Button variant="outlined" onClick={add}>Add</Button>
            </Stack>
            <Box sx={{display: 'flex', flexWrap: 'wrap', gap: 1}}>
                {items.map(i => <Chip key={i} label={i} onDelete={() => save(items.filter(x => x !== i))} sx={{fontFamily: 'monospace'}}/>)}
                {!items.length && <Typography variant="body2" sx={{color: 'text.secondary'}}>Empty</Typography>}
            </Box>
        </Paper>
    );
}
