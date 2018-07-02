import React, { Component } from 'react';
import DeploymentList from './DeploymentList.js';
//import logo from './logo.svg';
//import './App.css';

class DeploymentOperator extends Component {
  render() {
    return (
      <div className="App">
        <header className="App-header">
          <h1 className="App-title">ArangoDeployments....</h1>
        </header>
        <DeploymentList/>
      </div>
    );
  }
}

export default DeploymentOperator;
