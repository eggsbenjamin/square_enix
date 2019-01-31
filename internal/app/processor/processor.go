package processor

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/eggsbenjamin/square_enix/internal/app/db"
	"github.com/eggsbenjamin/square_enix/internal/app/models"
	"github.com/eggsbenjamin/square_enix/internal/app/repository"
)

var ErrNoRunningProcess = errors.New("no running process")

type Processor interface {
	Process(batchSize int) error
}

type processor struct {
	db                 db.DB
	processRepoFactory repository.ProcessRepositoryFactory
	elementRepoFactory repository.ElementRepositoryFactory
}

func (p *processor) Process(batchSize int) error {
	// query db for running process
	runningProcesses, err := p.processRepoFactory.CreateProcessRepository(p.db).GetByStatus(models.PROCESS_STATUS_RUNNING)
	if err != nil {
		// if not found return meaningful error - ErrNoRunningProcess
		return errors.Wrap(err, "error retreiving running processes")
	}

	if len(runningProcesses) == 0 {
		return ErrNoRunningProcess
	}

	if len(runningProcesses) > 1 {
		return errors.New("more thasn one process running. Aborting") // this should never happen but cover the edge case.
	}

	process := runningProcesses[0]

	/*
		start a transaction and perform all following steps using it
			- any error should rollback the transaction
	*/

	tx, err := p.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "error beginning transaction")
	}

	processRepo := p.processRepoFactory.CreateProcessRepository(tx)
	elementRepo := p.elementRepoFactory.CreateElementRepository(tx)

	/*
		query element table for elements <= batchSize that:
			- have no entry in the ProcessElement table
			- are unlocked
			- were created on or before the created_at field of the current running process
	*/

	elementsToBeProcessed, err := elementRepo.LockElementsForUpdate(process.ID, 50)
	if err != nil {
		return errors.Wrap(err, "error locking elements")
	}

	if len(elementsToBeProcessed) == 0 {

		/*
			if no elements are found:
			 - query the ProcessElement table for the number of elements that have been processed during the current running process
			 - query the Element table for the total number of elements that:
				- were created on or before the created_at field of the current running process
			- if these numbers are equal then the current running process has no more elements to process
				- if there are more elements to process they're currently locked and being processed by another instance so return nil here as no error has occurred
				- if there are no more elements to process
					- lock the process table
					- update the current running process's status to COMPLETE and unlock the table
		*/

		processedElements, err := elementRepo.GetElementsByProcessID(process.ID)
		if err != nil {
			return errors.Wrap(err, "error retreiving processed elements")
		}

		elementsCreatedBeforeProcess, err := elementRepo.GetElementsCreatedBefore(process.CreatedAt)
		if err != nil {
			return errors.Wrap(err, "error retreiving elements to be processed")
		}

		if len(processedElements) < len(elementsCreatedBeforeProcess) {
			return nil // another instance has locked rows
		}

		process.Status = models.PROCESS_STATUS_COMPLETE

		if err := processRepo.UpdateProcess(process); err != nil {
			return errors.Wrap(err, "error completing process")
		}

		return tx.Commit()
	}

	/*
		if elements are found:
		- process all of the elements - convert data field to uppercase
		- persist all of the updated elements
		- commit the transaction
		- return nil
	*/

	for _, element := range elementsToBeProcessed {
		element.Data = strings.ToUpper(element.Data)

		if err := elementRepo.UpdateElement(element); err != nil {
			return errors.Wrap(err, "error updating element")
		}
	}

	// commit the transaction

	return tx.Commit()
}
