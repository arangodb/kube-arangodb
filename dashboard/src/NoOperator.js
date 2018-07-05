import React, { Component } from 'react';
import logo from './logo.svg';
import './App.css';

class NoOperator extends Component {
  render() {
    return (
        <div className="App">
          <header className="App-header">
            <img src={logo} className="App-logo" alt="logo" />
            <h1 className="App-title">Welcome to Kube-ArangoDB</h1>
          </header>
          <p className="App-intro">
            There are no operators available yet.
          </p>
          {this.props["pod-info"]}
        </div>
    );
  }
}

export default NoOperator;
