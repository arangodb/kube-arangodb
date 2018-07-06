import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import App from './App';
import Auth from './auth/Auth.js';
import registerServiceWorker from './registerServiceWorker';

ReactDOM.render(<Auth><App /></Auth>, document.getElementById('root'));
registerServiceWorker();
