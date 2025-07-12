// Main entry point for the Tournament Planner React application
// This file is the starting point that React looks for when the app loads

import React from 'react';
import ReactDOM from 'react-dom/client';
import './styles/index.css';
import App from './App';

// Find the root element in our HTML file
const root = ReactDOM.createRoot(document.getElementById('root'));

// Render our main App component into the root element
// StrictMode helps detect potential problems during development
root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);

// Optional: Measure performance in your app
// You can pass a function to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
// import reportWebVitals from './reportWebVitals';
// reportWebVitals();