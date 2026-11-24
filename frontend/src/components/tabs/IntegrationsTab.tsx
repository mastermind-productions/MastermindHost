import {useEffect, useState} from 'react';
import {Button, Chip, FormControlLabel, Paper, Stack, Switch, TextField, Typography} from '@mui/material';
import {backend, Host, run} from '../../api';
import type {TabProps} from '../ServerPage';

export default function IntegrationsTab({server, onChanged}: TabProps) {
    const [discord, setDiscord] = useState<backend.DiscordHook>(new backend.DiscordHook({url: '', enabled: false}));
    const [branch, setBranch] = useState('');
    const [betaPass, setBetaPass] = useState('');
    const [fw, setFw] = useState<boolean | null>(null);

    const loadFw = () => Host.FirewallStatus(server.id).then(setFw);
    useEffect(() => {
        Host.GetServerRaw(server.id).then(s => { setDiscord(s.discord); setBranch(s.betaBranch); setBetaPass(s.betaPassword); });
        loadFw();
    }, [server.id]);

    return (
        <Stack spacing={2}>
            <Paper sx={{p: 2}}>
                <Typography variant="subtitle1" sx={{fontWeight: 600, mb: 1}}>Discord webhook</Typography>
                <Typography variant="body2" sx={{color: 'text.secondary', mb: 2}}>Posts a message when the server goes online, offline, starts, restarts, updates or shuts down.</Typography>
                <Stack spacing={1.5}>
                    <TextField fullWidth label="Webhook URL" type="password" value={discord.url} onChange={e => setDiscord(new backend.DiscordHook({...discord, url: e.target.value}))}/>
                    <Stack direction="row" spacing={1} sx={{alignItems: 'center'}}>
                        <FormControlLabel label="Enabled" control={<Switch checked={discord.enabled} onChange={e => setDiscord(new backend.DiscordHook({...discord, enabled: e.target.checked}))}/>}/>
                        <Stack direction="row" spacing={1} sx={{flex: 1, justifyContent: 'flex-end'}}>
                            <Button variant="outlined" onClick={() => run(Host.SetDiscord(server.id, discord).then(() => Host.TestDiscord(server.id)), 'Test message sent')}>Save & test</Button>
                            <Button variant="contained" onClick={() => run(Host.SetDiscord(server.id, discord), 'Discord settings saved')}>Save</Button>
                        </Stack>
                    </Stack>
                </Stack>
            </Paper>
            <Paper sx={{p: 2}}>
                <Typography variant="subtitle1" sx={{fontWeight: 600, mb: 1}}>Steam beta branch</Typography>
                <Typography variant="body2" sx={{color: 'text.secondary', mb: 2}}>Leave empty for the public branch. Takes effect on the next install / update.</Typography>
                <Stack direction="row" spacing={1}>
                    <TextField label="Branch name" value={branch} onChange={e => setBranch(e.target.value)} sx={{flex: 1}}/>
                    <TextField label="Branch password" type="password" value={betaPass} onChange={e => setBetaPass(e.target.value)} sx={{flex: 1}}/>
                    <Button variant="contained" onClick={() => run(Host.SetBeta(server.id, branch, betaPass), 'Branch saved').then(onChanged)}>Save</Button>
                </Stack>
            </Paper>
            <Paper sx={{p: 2}}>
                <Stack direction="row" spacing={1} sx={{alignItems: 'center', mb: 1}}>
                    <Typography variant="subtitle1" sx={{fontWeight: 600}}>Windows Firewall</Typography>
                    {fw !== null && <Chip size="small" color={fw ? 'success' : 'default'} label={fw ? 'Rule present' : 'No rule'}/>}
                </Stack>
                <Typography variant="body2" sx={{color: 'text.secondary', mb: 2}}>
                    Adds inbound UDP rules for port {server.port}-{server.port + 1} and the server executable. Requires a UAC prompt. RCON is not exposed.
                </Typography>
                <Stack direction="row" spacing={1}>
                    <Button variant="contained" onClick={() => run(Host.OpenFirewall(server.id), 'Firewall rules added').then(() => setTimeout(loadFw, 1500))}>Open ports</Button>
                    <Button variant="outlined" color="error" onClick={() => run(Host.CloseFirewall(server.id), 'Firewall rules removed').then(() => setTimeout(loadFw, 1500))}>Remove rules</Button>
                    <Button onClick={loadFw}>Refresh</Button>
                </Stack>
            </Paper>
        </Stack>
    );
}
