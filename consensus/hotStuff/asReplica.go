/*
 * Copyright 2020 The SealABC Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

/*
 * Copyright 2020 The SealABC Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package hotStuff

import (
    "github.com/SealSC/SealABC/consensus"
    "github.com/SealSC/SealABC/log"
    "github.com/SealSC/SealABC/metadata/message"
)

func (b *basicService) gotPrepare(consensusData SignedConsensusData) (reply *message.Message) {
    if !b.verifyPrepareMessage(consensusData) {
        return
    }

    if b.currentView < consensusData.ViewNumber {
        b.currentView = consensusData.ViewNumber
        log.Log.Warn("local view is lower then network view, set view to network view")
    }

    voteMsg, err := b.buildVoteMessage(consensusData.Phase, consensusData.Payload)
    if err != nil {
        log.Log.Error("build vote message failed")
        return
    }

    b.currentPhase = consensusPhases.Prepare
    b.viewChangeTrigger.Reset(b.config.ConsensusTimeout)
    b.sendMessageToLeader(voteMsg)
    return
}

func (b *basicService)processCommonPhaseMessage(consensusData ConsensusData) {
    allPhases :=  consensusPhases
    switch consensusData.Phase {
    case allPhases.PreCommit.String():
        b.prepareQC = &consensusData.Justify
        b.currentPhase = allPhases.PreCommit

    case allPhases.Commit.String():
        b.lockedQC = &consensusData.Justify
        b.currentPhase = allPhases.Commit

    case allPhases.Decide.String():
        b.currentPhase = allPhases.Decide
    }
}

func (b *basicService) gotCommonPhaseMessage(consensusData SignedConsensusData) (reply *message.Message) {
    if b.isCurrentLeader() {
        return
    }

    validPhase := b.verifyPhase(consensusData.ConsensusData)
    if !validPhase {
        return
    }

    b.processCommonPhaseMessage(consensusData.ConsensusData)

    if b.currentPhase == consensusPhases.Decide {
        if b.externalProcessor != nil {
            b.externalProcessor.EventProcessor(consensus.Event.Success, consensusData.Justify.Payload.CustomerData)
        }

        b.currentPhase = consensusPhases.NewView
        b.currentView += 1
        b.newRound()

        //log.Log.Println("consensus success! need send new view to next leader @view ", b.currentView)

        return
    }

    //log.Log.Println("common phase verify success, start build vote message in phase ", consensusData.Phase)
    voteMsg, err := b.buildVoteMessage(consensusData.Phase, consensusData.Justify.Payload)
    if err != nil {
        log.Log.Error("build vote message failed")
        return
    }
    b.viewChangeTrigger.Reset(b.config.ConsensusTimeout)

    //log.Log.Println("build vote message in phase ", consensusData.Phase, " over")

    b.sendMessageToLeader(voteMsg)
    return
}


func (b *basicService) registerReplicaProcessor() {
    b.consensusProcessor[messageTypes.Prepare.String()] = b.gotPrepare
    b.consensusProcessor[messageTypes.PreCommit.String()] = b.gotCommonPhaseMessage
    b.consensusProcessor[messageTypes.Commit.String()] = b.gotCommonPhaseMessage
    b.consensusProcessor[messageTypes.Decide.String()] = b.gotCommonPhaseMessage
}