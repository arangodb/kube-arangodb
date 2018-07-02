import React, { Component } from 'react';
import DeploymentOperator from './deployment/DeploymentOperator.js';
import { apiGet } from './api/api.js';
import logo from './logo.svg';
import './App.css';

class App extends Component {
  constructor() {
    super();
    this.state = {operators:{}};
  }

  async componentDidMount() {
    this.reloadOperators();
  }

  async reloadOperators() {
    const operators = await apiGet('/api/operators');
    this.setState({operators});
  }

  render() {
    setTimeout(this.reloadOperators.bind(this), 5000);
    if (this.state.operators.deployment) {
      return (<DeploymentOperator/>);
    }
    return (
      <div className="App">
        <header className="App-header">
          <img src={logo} className="App-logo" alt="logo" />
          <h1 className="App-title">Welcome to Kube-ArangoDB</h1>
        </header>
        <p className="App-intro">
          There are no operators available yet.
        </p>
      </div>
    );
  }
}

export default App;
