import {ComponentType, useState} from 'react';
import {Box, Chip, IconButton, LinearProgress, Paper, Stack, Tab, Tabs, Typography} from '@mui/material';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import {backend, statusColor} from '../api';
import Actions from './Actions';
import OverviewTab from './tabs/OverviewTab';
import CommandLineTab from './tabs/CommandLineTab';
import GameSettingsTab from './tabs/GameSettingsTab';
import AdvancedConfigTab from './tabs/AdvancedConfigTab';
import ConfigFilesTab from './tabs/ConfigFilesTab';
import ModsTab from './tabs/ModsTab';
import PlayersTab from './tabs/PlayersTab';
import ListsTab from './tabs/ListsTab';
import BackupsTab from './tabs/BackupsTab';
import SchedulerTab from './tabs/SchedulerTab';
import IntegrationsTab from './tabs/IntegrationsTab';
import LogsTab from './tabs/LogsTab';
import EventsTab from './tabs/EventsTab';
import PerformanceTab from './tabs/PerformanceTab';
import NetworkTab from './tabs/NetworkTab';
import PalSettingsTab from './tabs/PalSettingsTab';
import ScriptOverviewTab from './tabs/ScriptOverviewTab';
import ScriptConfigTab from './tabs/ScriptConfigTab';

export type TabProps = {server: backend.ServerInfo; onChanged: () => void};

const TABS: [string, ComponentType<TabProps>][] = [
    ['Overview', OverviewTab],
    ['Command Line', CommandLineTab],
    ['Game Settings', GameSettingsTab],
    ['Advanced Config', AdvancedConfigTab],
    ['Config Files', ConfigFilesTab],
    ['Mods', ModsTab],
    ['Players & RCON', PlayersTab],
    ['Admins & Lists', ListsTab],
    ['Backups', BackupsTab],
    ['Performance', PerformanceTab],
    ['Ports & A2S', NetworkTab],
    ['Scheduler', SchedulerTab],
    ['Discord / Beta / Firewall', IntegrationsTab],
    ['Logs', LogsTab],
    ['Events', EventsTab],
];

const PAL_TABS: [string, ComponentType<TabProps>][] = TABS
    .filter(([label]) => label !== 'Mods')
    .map(([label, c]) => [label, label === 'Game Settings' ? PalSettingsTab : c]);

const SCRIPT_TABS: [string, ComponentType<TabProps>][] = [
    ['Overview', ScriptOverviewTab],
    ['Launch Config', ScriptConfigTab],
    ['Scheduler', SchedulerTab],
    ['Events', EventsTab],
];

export default function ServerPage({server, onBack, onChanged}: {server: backend.ServerInfo; onBack: () => void; onChanged: () => void}) {
    const [tab, setTab] = useState(0);
    const working = server.status === 'installing' || server.status === 'updating';
    const tabs = server.script ? SCRIPT_TABS : server.game === 'palworld' ? PAL_TABS : TABS;
    const Comp = tabs[Math.min(tab, tabs.length - 1)][1];
    return (
        <Box sx={{display: 'flex', flexDirection: 'column', gap: 2, height: '100%'}}>
            <Paper sx={{p: 2}}>
                <Stack direction="row" spacing={2} sx={{alignItems: 'center', flexWrap: 'wrap', rowGap: 1}}>
                    <IconButton onClick={onBack}><ArrowBackIcon/></IconButton>
                    <Box sx={{minWidth: 0}}>
                        <Typography variant="h5" sx={{fontWeight: 700}} noWrap>{server.name}</Typography>
                        <Typography variant="body2" sx={{color: 'text.secondary'}} noWrap>{server.gameName} - {server.dir}</Typography>
                    </Box>
                    <Chip color={statusColor(server.status)} label={server.status.toUpperCase()}/>
                    <Box sx={{flex: 1}}/>
                    <Actions server={server} onChanged={onChanged}/>
                </Stack>
                {working && <LinearProgress sx={{mt: 1.5}} variant={server.progress > 0 ? 'determinate' : 'indeterminate'} value={server.progress}/>}
            </Paper>
            <Paper sx={{flex: 1, display: 'flex', minHeight: 0}}>
                <Tabs orientation="vertical" value={tab} onChange={(_, v) => setTab(v)} variant="scrollable"
                      sx={{borderRight: 1, borderColor: 'divider', minWidth: 210, '& .MuiTab-root': {alignItems: 'flex-start'}}}>
                    {tabs.map(([label]) => <Tab key={label} label={label}/>)}
                </Tabs>
                <Box sx={{flex: 1, p: 2, overflow: 'auto', minWidth: 0}}>
                    <Comp key={server.id + tab} server={server} onChanged={onChanged}/>
                </Box>
            </Paper>
        </Box>
    );
}
