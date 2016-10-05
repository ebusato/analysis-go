
void plot(TString name, int evt)
{
	gStyle->SetPadGridX(true);
	gStyle->SetPadGridY(true);
	
	TFile* f = new TFile(name, "read");
	TTree* t = (TTree*) f->Get("tree");
	
	TCanvas* c1 = new TCanvas("c1","c1",900,900);
	c1->Divide(2,2);
	c1->cd(1);
	tree->Draw("T30[0] - T30[1]");
	c1->cd(2);
	tree->Draw("Pulse[1] : SampleTimes", Form("Evt == %i", evt),"goff");
	TGraph *g1 = new TGraph(999,tree->GetV2(),tree->GetV1());
	g1->SetMarkerColor(kRed);
	g1->SetMarkerStyle(8);
	g1->SetMarkerSize(0.5);
	tree->Draw("Pulse[0]*Ampl[1]/Ampl[0] : SampleTimes", Form("Evt == %i", evt),"same");
	TGraph *g2 = new TGraph(999,tree->GetV2(),tree->GetV1());
	g2->SetMarkerColor(kBlue);
	g2->SetMarkerStyle(8);
	g2->SetMarkerSize(0.5);
	TMultiGraph* mg = new TMultiGraph();
	mg->Add(g1);
	mg->Add(g2);
	mg->Draw("ap");
	tree->Draw("Ampl[1]", Form("Evt == %i", evt),"goff");
	amplitude1 = *tree->GetV1();
	tree->Draw("Ampl[0]", Form("Evt == %i", evt),"goff");
	amplitude0 = *tree->GetV1();
	TLine* l = new TLine(0, amplitude1*0.3, 200, amplitude1*0.3);
	l->SetLineWidth(2);
	l->Draw("same");
	tree->Draw("T30[0] - T30[1]", Form("Evt == %i", evt),"goff");
	TLatex* la1 = new TLatex(110, amplitude1 * 0.6, Form("#Delta t = %f ns", *tree->GetV1()));
	la1->Draw();
	TLatex* la2 = new TLatex(110, amplitude1 * 0.8, Form("ampl[0]=%f", amplitude0));
	la2->SetTextColor(kBlue);
	la2->Draw();
	TLatex* la3 = new TLatex(110, amplitude1 * 0.7, Form("ampl[1]=%f", amplitude1));
	la3->SetTextColor(kRed);
	la3->Draw();
}