package processor

import (
	"log"
	"strings"

	"github.com/pkg/errors"

	"github.com/eggsbenjamin/square_enix/internal/app/db"
	"github.com/eggsbenjamin/square_enix/internal/app/models"
	"github.com/eggsbenjamin/square_enix/internal/app/repository"
)

var (
	ErrNoProcessExists        = errors.New("no process exists")
	ErrNoRunningProcessExists = errors.New("no running process")
	ErrRunningProcessExists   = errors.New("running process exists")
)

type Processor interface {
	Start() error
	Pause() error
	RunningProcessExists() (bool, error)
	ProcessBatch(batchSize int) error
	GetLatestsStat() (int, error)
}

type processor struct {
	db                 db.DB
	processRepoFactory repository.ProcessRepositoryFactory
	elementRepoFactory repository.ElementRepositoryFactory
}

func NewProcessor(
	db db.DB,
	processRepoFactory repository.ProcessRepositoryFactory,
	elementRepoFactory repository.ElementRepositoryFactory,
) Processor {
	return &processor{
		db:                 db,
		processRepoFactory: processRepoFactory,
		elementRepoFactory: elementRepoFactory,
	}
}

func (p *processor) Start() error {
	tx, err := p.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "error beginning transaction")
	}

	pausedProcesses, err := p.processRepoFactory.CreateProcessRepository(p.db).GetByStatus(models.PROCESS_STATUS_PAUSED)
	if err != nil {
		return errors.Wrap(err, "error retreiving running processes")
	}

	if len(pausedProcesses) > 0 {
		pausedProcess := pausedProcesses[0]
		pausedProcess.Status = models.PROCESS_STATUS_RUNNING

		log.Printf("resuming process: %d\n", pausedProcess.ID)
		return p.processRepoFactory.CreateProcessRepository(p.db).UpdateProcess(pausedProcess)
	}

	if _, err := p.processRepoFactory.CreateProcessRepository(tx).CreateNewProcess(); err != nil {
		if err == repository.ErrRunningProcessExists {
			return ErrRunningProcessExists
		}
		return err
	}

	return tx.Commit()
}

func (p *processor) Pause() error {
	processRepo := p.processRepoFactory.CreateProcessRepository(p.db)
	latestProcess, err := processRepo.GetLatestProcess()
	if err != nil {
		if err == repository.ErrNoProcessExists {
			return ErrNoProcessExists
		}
		return err
	}

	if latestProcess.Status != models.PROCESS_STATUS_RUNNING {
		return ErrNoRunningProcessExists
	}

	latestProcess.Status = models.PROCESS_STATUS_PAUSED

	return processRepo.UpdateProcess(latestProcess)
}

func (p *processor) RunningProcessExists() (bool, error) {
	runningProcesses, err := p.processRepoFactory.CreateProcessRepository(p.db).GetByStatus(models.PROCESS_STATUS_RUNNING)
	if err != nil {
		return false, errors.Wrap(err, "error retreiving running processes")
	}

	return len(runningProcesses) > 0, nil
}

func (p *processor) ProcessBatch(batchSize int) error {
	// query db for running process
	runningProcesses, err := p.processRepoFactory.CreateProcessRepository(p.db).GetByStatus(models.PROCESS_STATUS_RUNNING)
	if err != nil {
		return errors.Wrap(err, "error retreiving running processes")
	}

	if len(runningProcesses) == 0 {
		return ErrNoRunningProcessExists
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

	elementsToBeProcessed, err := elementRepo.LockElementsForUpdate(process.ID, batchSize)
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

		log.Printf("completing proces: %d\n", process.ID)
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

	log.Printf("processing %d elements as part of process: %d\n", len(elementsToBeProcessed), process.ID)

	for _, element := range elementsToBeProcessed {
		element.Data = strings.ToUpper(element.Data)

		if err := elementRepo.UpdateElementForProcess(element, process.ID); err != nil {
			return errors.Wrap(err, "error updating element")
		}
	}

	// commit the transaction

	return tx.Commit()
}

func (p *processor) GetLatestsStat() (int, error) {
	latestProcess, err := p.processRepoFactory.CreateProcessRepository(p.db).GetLatestProcess()
	if err != nil {
		if err == repository.ErrNoProcessExists {
			return 0, ErrNoProcessExists
		}

		return 0, err
	}

	processElements, err := p.elementRepoFactory.CreateElementRepository(p.db).GetElementsByProcessID(latestProcess.ID)
	if err != nil {
		return 0, err
	}

	return len(processElements), nil
}
