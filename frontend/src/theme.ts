import {createTheme} from '@mui/material/styles';

export const GREEN = '#22c55e';

export const theme = createTheme({
    palette: {
        mode: 'dark',
        primary: {main: GREEN, dark: '#15803d', light: '#4ade80', contrastText: '#03140a'},
        secondary: {main: '#0f3d24'},
        background: {default: '#050806', paper: '#0a120d'},
        divider: 'rgba(34,197,94,0.18)',
        text: {primary: '#e6f4ea', secondary: '#8fb39a'},
        success: {main: GREEN},
    },
    shape: {borderRadius: 8},
    typography: {
        fontFamily: '"Segoe UI", "Nunito", Roboto, sans-serif',
        button: {textTransform: 'none', fontWeight: 600},
    },
    components: {
        MuiPaper: {styleOverrides: {root: {backgroundImage: 'none', border: '1px solid rgba(34,197,94,0.14)'}}},
        MuiAppBar: {styleOverrides: {root: {background: '#030504', borderBottom: '1px solid rgba(34,197,94,0.25)'}}},
        MuiTab: {styleOverrides: {root: {textTransform: 'none', minHeight: 44}}},
        MuiTextField: {defaultProps: {size: 'small'}},
        MuiSelect: {defaultProps: {size: 'small'}},
        MuiButton: {defaultProps: {size: 'small'}},
        MuiTableCell: {styleOverrides: {root: {borderColor: 'rgba(34,197,94,0.12)'}}},
    },
});
