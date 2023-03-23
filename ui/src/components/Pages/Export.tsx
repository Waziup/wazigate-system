import React, { Component } from "react";

import {
    MDBBtn,
    MDBIcon,
    MDBCardText,
    MDBRow,
    MDBContainer,
    MDBCol
  } from "mdbreact";

export interface State {
  fromTime: Date;
  toTime: Date;
  duration: number;
  checkBins: boolean;
}

class PageExport extends React.Component<{},State> {


  constructor(props:{}) {
    super(props);
    var fromTime = new Date();
    fromTime.setDate(fromTime.getDate()-7);
    this.state= {
      fromTime: fromTime,
      toTime: new Date(),
      duration: 10,
      checkBins: false,
    }
  }
  convTime = (date: Date) => {
    //console.log("convTime_Date: " + date)
    return `${date.getFullYear()}-${padZero(date.getMonth()+1)}-${padZero(date.getDate())}T${padZero(date.getHours())}:${padZero(date.getMinutes())}`
  }
  
  convTimeWithSec = (date: Date) => {
    //console.log("convTime_Date: " + date)
    return `${date.getFullYear()}-${padZero(date.getMonth()+1)}-${padZero(date.getDate())}T${padZero(date.getHours())}:${padZero(date.getMinutes())}:${padZero(date.getSeconds())}`
  }

  render() {
      return (
          <>
            <h4>Export Section</h4>
            <br/>
            <MDBCardText>In this section you are able to download the gateways data of all devices at once. You can use to perform a backup, to have all data in one place and for machine learning applications. </MDBCardText>
            <br/>
            <MDBContainer>
            <MDBRow>
              <MDBCol>
                  <MDBCardText>You can export the data of all sensors and actuators to a tree of CSV files:</MDBCardText>
                </MDBCol>
                <MDBCol>
                  <MDBBtn disabled={false} href="../../../exporttree">
                        <MDBIcon
                          icon="account_tree"
                          className="ml-2"
                          size="1x"
                        />{" "}
                        Export data to tree of CSV files
                  </MDBBtn>
                </MDBCol>
              </MDBRow>
              <br/>
              <MDBRow>
                <MDBCol>
                  <MDBCardText>You can export the data of all sensors and actuators to one CSV file:</MDBCardText>
                </MDBCol>
                <MDBCol>
                  <MDBBtn disabled={false} href="../../../exportall">
                        <MDBIcon
                          icon="all_match"
                          className="ml-2"
                          size="1x"
                        />{" "}
                        Export data to one CSV file
                  </MDBBtn>
                </MDBCol>
              </MDBRow>
              <br/>
              <MDBRow>
                <MDBCol>
                  <MDBCardText>You can export the data of all sensors and actuators to one CSV file. Additionally it also includes custom timespans and all data can be summarized in time bins. This is perfect for machine learning applications:</MDBCardText>
                </MDBCol>
                <MDBCol>
                  <MDBCardText>From: </MDBCardText>
                  <input type="datetime-local" id="from-time"
                      name="from-time" value={this.convTime(this.state.fromTime)}
                      min="1990-01-01T00:00"
                      onChange={(ev) => {
                        this.setState({
                          fromTime: new Date(ev.currentTarget.value)
                        })
                      }}>
                      </input>
                  <MDBCardText>To: </MDBCardText>
                  <input type="datetime-local" id="to-time"
                      name="to-time" defaultValue={this.convTime(this.state.toTime)}
                      min="1990-01-01T00:00"
                      onChange={(ev) => {
                        this.setState({
                          toTime: new Date(ev.currentTarget.value)
                        })
                      }}>
                      </input>
                  <MDBCardText>Bin Size in minutes: </MDBCardText>
                  <input type="number" id="bins-time" name="bins-time" value={this.state.duration}
                    onChange={(ev) => {
                          this.setState({
                            duration: parseInt(ev.currentTarget.value)
                          })
                        }}>
                  </input>
                  <MDBCardText>Omit diviating values (20%) inbetween bins : </MDBCardText>
                  <input type="checkbox" id="clear" name="clear"
                    onChange={(ev) => {
                      this.setState({
                        checkBins: ev.currentTarget.checked
                      })
                    }}>
                  </input>
                  <MDBBtn disabled={false} href={"../../../exportbins?from="+this.state.fromTime.toISOString()+"&to="+this.state.toTime.toISOString()+"&duration="+this.state.duration.toString()+"m"+"&check="+this.state.checkBins}>
                        <MDBIcon
                          icon="account_tree"
                          className="ml-2"
                          size="1x"
                        />{" "}
                        Export data to one CSV file, custom timespan, in bins
                  </MDBBtn>
                </MDBCol>
              </MDBRow>
            </MDBContainer>
          </>
      );
  }   
}

function padZero(t: number): string {
  if (t < 10) return "0"+t;
  return ""+t;
}

export default PageExport