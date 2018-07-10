import { Dimmer, Loader, Segment } from 'semantic-ui-react';
import React from 'react';

const Loading = ({message}) => (
        <Segment>
        <Dimmer inverted active>
          <Loader inverted>{message || "Loading..."}</Loader>
        </Dimmer>
        <div style={{minHeight:"3em"}}/>
      </Segment>
      );

export default Loading;
