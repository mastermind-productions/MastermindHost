import {useEffect, useState} from 'react';
import {Button, MenuItem, Stack, TextField, Typography} from '@mui/material';
import {backend, Host, run} from '../../api';
import type {TabProps} from '../ServerPage';

export default function ConfigFilesTab({server, onChanged}: TabProps) {
    const [files, setFiles] = useState<backend.ConfigFileDef[]>([]);
    const [key, setKey] = useState('gusIni');
    const [content, setContent] = useState('');
    const [orig, setOrig] = useState('');

    useEffect(() => { Host.GetConfigFiles(server.id).then(setFiles); }, [server.id]);
    const load = () => Host.ReadConfig(server.id, key).then(c => { setContent(c); setOrig(c); });
    useEffect(() => { load(); }, [server.id, key]);

    const def = files.find(f => f.key === key);
    const dirty = content !== orig;

    return (
        <Stack spacing={1.5} sx={{height: '100%'}}>
            <Stack direction="row" spacing={2} sx={{alignItems: 'center'}}>
                <TextField select label="File" value={key} onChange={e => setKey(e.target.value)} sx={{minWidth: 320}}>
                    {files.map(f => <MenuItem key={f.key} value={f.key}>{f.name}</MenuItem>)}
                </TextField>
                <Typography variant="body2" sx={{color: 'text.secondary', flex: 1}} noWrap>
                    {def?.description}{def?.relPath ? ` - ${def.relPath}` : ''}
                </Typography>
                <Button onClick={load} disabled={!dirty}>Revert</Button>
                <Button variant="contained" disabled={!dirty}
                        onClick={() => run(Host.WriteConfig(server.id, key, content), 'Saved').then(() => { load(); onChanged(); })}>Save</Button>
            </Stack>
            <TextField multiline fullWidth minRows={28} value={content} onChange={e => setContent(e.target.value)} spellCheck={false}
                       sx={{'& textarea': {fontFamily: 'Consolas, monospace', fontSize: 13, lineHeight: 1.45}}}/>
        </Stack>
    );
}
