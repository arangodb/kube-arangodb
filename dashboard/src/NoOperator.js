import React, { Component } from 'react';
import logo from './logo.svg';
import './App.css';
import { Message } from 'semantic-ui-react';

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
          {this.props.podInfoView}
          {(this.props.error) ? <Message error content={this.props.error}/> : null}
        </div>
    );
  }
}

export default NoOperator;
