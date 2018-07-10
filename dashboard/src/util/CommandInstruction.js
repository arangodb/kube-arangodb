import { Button, Modal, Segment } from 'semantic-ui-react';
import { CopyToClipboard } from 'react-copy-to-clipboard';
import React, { Component } from 'react';

class CommandInstruction extends Component {
  state = {open:false};

  close = () => { this.setState({open:false}); }
  open = () => { this.setState({open:true}); }

  render() {
    return (
      <Modal trigger={this.props.trigger} onClose={this.close} onOpen={this.open} open={this.state.open}>
        <Modal.Header>{this.props.title}</Modal.Header>
        <Modal.Content>
          <Modal.Description>
            <p>
              {this.props.description}
            </p>
            <Segment clearing>
              <code>{this.props.command}</code>
            </Segment>
          </Modal.Description>
        </Modal.Content>
        <Modal.Actions>
          <CopyToClipboard text={this.props.command} onCopy={this.close}>
            <Button
              positive
              icon='copy'
              labelPosition='right'
              content="Copy"
            />
          </CopyToClipboard>
        </Modal.Actions>
      </Modal>
    );
  }
}

export default CommandInstruction;
