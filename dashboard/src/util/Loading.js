import React, { Component } from 'react';
import { Dimmer, Loader, Segment } from 'semantic-ui-react';

class Loading extends Component {
  render() {
    return (
        <Segment>
        <Dimmer inverted active>
          <Loader inverted>{this.props.message || "Loading..."}</Loader>
        </Dimmer>
        <div style={{"min-height":"3em"}}/>
      </Segment>
      );
  }
}

export default Loading;
